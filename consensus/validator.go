// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package consensus

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/hotstuff"
	"github.com/aungmawjj/juria-blockchain/logger"
)

type validator struct {
	resources *Resources
	state     *state
	hotstuff  *hotstuff.Hotstuff

	mtxProposal sync.Mutex
}

func (vld *validator) start() {
	go vld.proposalLoop()
	go vld.voteLoop()
	go vld.newViewLoop()
	logger.I().Info("started validator")
}

func (vld *validator) proposalLoop() {
	sub := vld.resources.MsgSvc.SubscribeProposal(100)
	defer sub.Unsubscribe()

	for e := range sub.Events() {
		if err := vld.onReceiveProposal(e.(*core.Block)); err != nil {
			logger.I().Warnw("on received proposal failed", "error", err)
		}
	}
}

func (vld *validator) voteLoop() {
	sub := vld.resources.MsgSvc.SubscribeVote(1000)
	defer sub.Unsubscribe()

	for e := range sub.Events() {
		if err := vld.onReceiveVote(e.(*core.Vote)); err != nil {
			logger.I().Warnw("received vote failed", "error", err)
		}
	}
}

func (vld *validator) newViewLoop() {
	sub := vld.resources.MsgSvc.SubscribeNewView(100)
	defer sub.Unsubscribe()

	for e := range sub.Events() {
		if err := vld.onReceiveNewView(e.(*core.QuorumCert)); err != nil {
			logger.I().Warnw("received new view failed", "error", err)
		}
	}
}

func (vld *validator) onReceiveProposal(proposal *core.Block) error {
	vld.mtxProposal.Lock()
	defer vld.mtxProposal.Unlock()

	if err := proposal.Validate(vld.resources.VldStore); err != nil {
		return err
	}
	if err := vld.syncParentBlocks(proposal.Proposer(), proposal); err != nil {
		return err
	}
	if err := vld.resources.TxPool.SyncTxs(proposal.Proposer(), proposal.Transactions()); err != nil {
		return err
	}
	vld.state.setBlock(proposal)
	return vld.updateHotstuff(proposal, true)
}

func (vld *validator) syncParentBlocks(peer *core.PublicKey, blk *core.Block) error {
	parent := vld.state.getBlockOnLocalNode(blk.ParentHash())
	if parent != nil {
		if blk.Height() != parent.Height()+1 {
			return fmt.Errorf("invalid block height")
		}
		return nil
	}
	parent, err := vld.requestBlock(peer, blk.ParentHash())
	if err != nil {
		return err
	}
	if blk.Height() != parent.Height()+1 {
		return fmt.Errorf("invalid block height")
	}
	if err := vld.syncParentBlocks(peer, parent); err != nil { // sync parent bocks recursive
		return err
	}
	if err := vld.resources.TxPool.SyncTxs(peer, parent.Transactions()); err != nil {
		return err
	}
	vld.state.setBlock(parent)
	return vld.updateHotstuff(parent, false)
}

func (vld *validator) requestBlock(peer *core.PublicKey, hash []byte) (*core.Block, error) {
	blk, err := vld.resources.MsgSvc.RequestBlock(peer, hash)
	if err != nil {
		return nil, fmt.Errorf("request block error %w", err)
	}
	if err := blk.Validate(vld.resources.VldStore); err != nil {
		return nil, fmt.Errorf("validate block error %w", err)
	}
	return blk, nil
}

func (vld *validator) updateHotstuff(blk *core.Block, voting bool) error {
	vld.state.mtxUpdate.Lock()
	defer vld.state.mtxUpdate.Unlock()

	if !voting {
		vld.hotstuff.Update(newHsBlock(blk, vld.state))
		return nil
	}
	if err := vld.canVoteProposal(blk); err != nil {
		vld.hotstuff.Update(newHsBlock(blk, vld.state))
		return err
	}
	vld.hotstuff.OnReceiveProposal(newHsBlock(blk, vld.state))
	return nil
}

func (vld *validator) canVoteProposal(proposal *core.Block) error {
	if !vld.state.isLeader(proposal.Proposer()) {
		return fmt.Errorf("proposer is not leader")
	}
	bh := vld.hotstuff.GetBExec().Height()
	if bh != proposal.ExecHeight() {
		return fmt.Errorf("invalid exec height")
	}
	mr := vld.resources.Storage.GetMerkleRoot()
	if !bytes.Equal(mr, proposal.MerkleRoot()) {
		return fmt.Errorf("invalid merkle root")
	}
	return vld.resources.TxPool.VerifyProposalTxs(proposal.Transactions())
}

func (vld *validator) onReceiveVote(vote *core.Vote) error {
	if err := vote.Validate(vld.resources.VldStore); err != nil {
		return err
	}
	vld.hotstuff.OnReceiveVote(newHsVote(vote, vld.state))
	return nil
}

func (vld *validator) onReceiveNewView(qc *core.QuorumCert) error {
	if err := qc.Validate(vld.resources.VldStore); err != nil {
		return err
	}
	vld.hotstuff.UpdateQCHigh(newHsQC(qc, vld.state))
	return nil
}
