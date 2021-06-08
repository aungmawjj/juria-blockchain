// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package consensus

import (
	"bytes"
	"encoding/base64"
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

	stopCh chan struct{}
}

func (vld *validator) start() {
	if vld.stopCh != nil {
		return
	}
	vld.stopCh = make(chan struct{})
	go vld.proposalLoop()
	go vld.voteLoop()
	go vld.newViewLoop()
	logger.I().Info("started validator")
}

func (vld *validator) stop() {
	if vld.stopCh == nil {
		return // not started yet
	}
	select {
	case <-vld.stopCh: // already stopped
		return
	default:
	}
	close(vld.stopCh)
	logger.I().Info("stopped validator")
	vld.stopCh = nil
}

func (vld *validator) proposalLoop() {
	sub := vld.resources.MsgSvc.SubscribeProposal(100)
	defer sub.Unsubscribe()

	for {
		select {
		case <-vld.stopCh:
			return

		case e := <-sub.Events():
			if err := vld.onReceiveProposal(e.(*core.Block)); err != nil {
				logger.I().Warnf("on received proposal failed, %+v", err)
			}
		}
	}
}

func (vld *validator) voteLoop() {
	sub := vld.resources.MsgSvc.SubscribeVote(1000)
	defer sub.Unsubscribe()

	for {
		select {
		case <-vld.stopCh:
			return

		case e := <-sub.Events():
			if err := vld.onReceiveVote(e.(*core.Vote)); err != nil {
				logger.I().Warnf("received vote failed, %+v", err)
			}
		}
	}
}

func (vld *validator) newViewLoop() {
	sub := vld.resources.MsgSvc.SubscribeNewView(100)
	defer sub.Unsubscribe()

	for {
		select {
		case <-vld.stopCh:
			return

		case e := <-sub.Events():
			if err := vld.onReceiveNewView(e.(*core.QuorumCert)); err != nil {
				logger.I().Warnf("received new view failed, %+v", err)
			}
		}
	}
}

func (vld *validator) onReceiveProposal(proposal *core.Block) error {
	vld.mtxProposal.Lock()
	defer vld.mtxProposal.Unlock()

	if err := proposal.Validate(vld.resources.VldStore); err != nil {
		return err
	}
	pidx := vld.resources.VldStore.GetValidatorIndex(proposal.Proposer())
	logger.I().Debugw("received proposal", "proposer", pidx, "height", proposal.Height())
	parent, err := vld.getParentBlock(proposal)
	if err != nil {
		return err
	}
	return vld.verifyWithParentAndUpdateHotstuff(
		proposal.Proposer(), proposal, parent, true)
}

func (vld *validator) getParentBlock(proposal *core.Block) (*core.Block, error) {
	parent := vld.state.getBlock(proposal.ParentHash())
	if parent != nil {
		return parent, nil
	}
	if err := vld.syncMissingCommitedBlocks(proposal); err != nil {
		return nil, err
	}
	return vld.syncMissingParentBlocksRecursive(proposal.Proposer(), proposal)
}

func (vld *validator) syncMissingCommitedBlocks(proposal *core.Block) error {
	commitHeight := vld.resources.Storage.GetBlockHeight()
	if proposal.ExecHeight() <= commitHeight {
		return nil // already sync commited blocks
	}
	// seems like I left behind. Lets check with qcRef to confirm
	// only qc is trusted and proposal is not
	if proposal.IsGenesis() {
		return fmt.Errorf("genesis block proposal")
	}
	qcRef := vld.state.getBlock(proposal.QuorumCert().BlockHash())
	if qcRef == nil {
		var err error
		qcRef, err = vld.requestBlock(
			proposal.Proposer(), proposal.QuorumCert().BlockHash())
		if err != nil {
			return err
		}
	}
	if qcRef.Height() < commitHeight {
		return fmt.Errorf("old qc ref %d", qcRef.Height())
	}
	return vld.syncForwardCommitedBlocks(
		proposal.Proposer(), commitHeight+1, proposal.ExecHeight())
}

func (vld *validator) syncForwardCommitedBlocks(peer *core.PublicKey, start, end uint64) error {
	var blk *core.Block
	for height := start; height < end; height++ { // end is exclusive
		var err error
		blk, err = vld.requestBlockByHeight(peer, height)
		if err != nil {
			return err
		}
		parent := vld.state.getBlock(blk.ParentHash())
		if parent == nil {
			return fmt.Errorf("cannot connect chain, parent not found")
		}
		err = vld.verifyWithParentAndUpdateHotstuff(peer, blk, parent, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func (vld *validator) syncMissingParentBlocksRecursive(
	peer *core.PublicKey, blk *core.Block,
) (*core.Block, error) {
	parent := vld.state.getBlock(blk.ParentHash())
	if parent != nil {
		return parent, nil // not missing
	}
	parent, err := vld.requestBlock(peer, blk.ParentHash())
	if err != nil {
		return nil, err
	}
	grandParent, err := vld.syncMissingParentBlocksRecursive(peer, parent)
	if err != nil {
		return nil, err
	}
	err = vld.verifyWithParentAndUpdateHotstuff(peer, parent, grandParent, false)
	if err != nil {
		return nil, err
	}
	return parent, nil
}

func (vld *validator) requestBlock(peer *core.PublicKey, hash []byte) (*core.Block, error) {
	blk, err := vld.resources.MsgSvc.RequestBlock(peer, hash)
	if err != nil {
		return nil, fmt.Errorf("cannot request block %w", err)
	}
	if err := blk.Validate(vld.resources.VldStore); err != nil {
		return nil, fmt.Errorf("validate block error %w", err)
	}
	return blk, nil
}

func (vld *validator) requestBlockByHeight(peer *core.PublicKey, height uint64) (*core.Block, error) {
	blk, err := vld.resources.MsgSvc.RequestBlockByHeight(peer, height)
	if err != nil {
		return nil, fmt.Errorf("cannot get block by height %d, %w", height, err)
	}
	if err := blk.Validate(vld.resources.VldStore); err != nil {
		return nil, fmt.Errorf("validate block error %w", err)
	}
	return blk, nil
}

func (vld *validator) verifyWithParentAndUpdateHotstuff(
	peer *core.PublicKey, blk, parent *core.Block, voting bool,
) error {
	if blk.Height() != parent.Height()+1 {
		return fmt.Errorf("invalid block height %d, parent %d",
			blk.Height(), parent.Height())
	}
	// must sync transactions before updating block to hotstuff
	if err := vld.resources.TxPool.SyncTxs(peer, blk.Transactions()); err != nil {
		return err
	}
	vld.state.setBlock(blk)
	return vld.updateHotstuff(blk, voting)
}

func (vld *validator) updateHotstuff(blk *core.Block, voting bool) error {
	vld.state.mtxUpdate.Lock()
	defer vld.state.mtxUpdate.Unlock()

	if !voting {
		vld.hotstuff.Update(newHsBlock(blk, vld.state))
		return nil
	}
	if err := vld.verifyProposalToVote(blk); err != nil {
		vld.hotstuff.Update(newHsBlock(blk, vld.state))
		return err
	}
	vld.hotstuff.OnReceiveProposal(newHsBlock(blk, vld.state))
	return nil
}

func (vld *validator) verifyProposalToVote(proposal *core.Block) error {
	if !vld.state.isLeader(proposal.Proposer()) {
		pidx := vld.resources.VldStore.GetValidatorIndex(proposal.Proposer())
		return fmt.Errorf("proposer %d is not leader", pidx)
	}
	// on node restart, not commited any blocks yet, don't check merkle root
	if vld.state.getCommitedHeight() != 0 {
		if err := vld.verifyMerkleRoot(proposal); err != nil {
			return err
		}
	}
	return vld.verifyProposalTxs(proposal)
}

func (vld *validator) verifyMerkleRoot(proposal *core.Block) error {
	bh := vld.resources.Storage.GetBlockHeight()
	if bh != proposal.ExecHeight() {
		return fmt.Errorf("invalid exec height")
	}
	mr := vld.resources.Storage.GetMerkleRoot()
	if !bytes.Equal(mr, proposal.MerkleRoot()) {
		return fmt.Errorf("invalid merkle root")
	}
	return nil
}

func (vld *validator) verifyProposalTxs(proposal *core.Block) error {
	for _, hash := range proposal.Transactions() {
		if vld.resources.Storage.HasTx(hash) {
			return fmt.Errorf("already commited tx: %s", base64String(hash))
		}
		tx := vld.resources.TxPool.GetTx(hash)
		if tx == nil {
			return fmt.Errorf("tx not found: %s", base64String(hash))
		}
		if tx.Expiry() != 0 && tx.Expiry() < proposal.Height() {
			return fmt.Errorf("expired tx: %s", base64String(hash))
		}
	}
	return nil
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

func base64String(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}
