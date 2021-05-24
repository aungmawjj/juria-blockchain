// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package consensus

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/logger"
	"golang.org/x/crypto/sha3"
)

type genesis struct {
	resources *Resources
	chainID   int64

	done chan struct{}

	// collect votes from all validators instead of majority for genesis block
	votes map[string]*core.Vote

	mtxVote    sync.Mutex
	mtxNewView sync.Mutex

	b0 atomic.Value
	q0 atomic.Value
}

func (gns *genesis) run() (*core.Block, *core.QuorumCert) {
	logger.Info("creating genesis block...")
	gns.done = make(chan struct{})
	gns.votes = make(map[string]*core.Vote, gns.resources.VldStore.ValidatorCount())

	go gns.proposalLoop()
	go gns.voteLoop()
	go gns.newViewLoop()
	gns.propose()

	<-gns.done
	return gns.getB0(), gns.getQ0()
}

func (gns *genesis) propose() {
	if !gns.isLeader(gns.resources.Signer.PublicKey()) {
		return
	}
	b0 := core.NewBlock().
		SetHeight(0).
		SetParentHash(hashChainID(gns.chainID)).
		SetTimestamp(time.Now().UnixNano()).
		Sign(gns.resources.Signer)

	gns.setB0(b0)
	logger.Info("created genesis block, broadcasting...")
	go gns.broadcastProposalLoop()
	gns.onReceiveVote(b0.ProposerVote())
}

func (gns *genesis) broadcastProposalLoop() {
	for {
		select {
		case <-gns.done:
			return
		default:
		}
		if gns.getQ0() == nil {
			if err := gns.resources.MsgSvc.BroadcastProposal(gns.getB0()); err != nil {
				logger.Error("broadcast proposal error", "error", err)
			}
		}
		time.Sleep(2 * time.Second)
	}
}

func (gns *genesis) isLeader(pubKey *core.PublicKey) bool {
	pidx, ok := gns.resources.VldStore.GetValidatorIndex(pubKey)
	if !ok {
		return false
	}
	return pidx == 0
}

func hashChainID(chainID int64) []byte {
	h := sha3.New256()
	binary.Write(h, binary.BigEndian, chainID)
	return h.Sum(nil)
}

func (gns *genesis) proposalLoop() {
	sub := gns.resources.MsgSvc.SubscribeProposal(10)
	defer sub.Unsubscribe()

	for {
		select {
		case <-gns.done:
			return

		case e := <-sub.Events():
			if err := gns.onReceiveProposal(e.(*core.Block)); err != nil {
				logger.Error("on receive proposal", "error", err)
			}
		}
	}
}

func (gns *genesis) voteLoop() {
	sub := gns.resources.MsgSvc.SubscribeVote(100)
	defer sub.Unsubscribe()

	for {
		select {
		case <-gns.done:
			return

		case e := <-sub.Events():
			if err := gns.onReceiveVote(e.(*core.Vote)); err != nil {
				logger.Error("on receive vote", "error", err)
			}
		}
	}
}

func (gns *genesis) newViewLoop() {
	sub := gns.resources.MsgSvc.SubscribeNewView(10)
	defer sub.Unsubscribe()

	for {
		select {
		case <-gns.done:
			return

		case e := <-sub.Events():
			if err := gns.onReceiveNewView(e.(*core.QuorumCert)); err != nil {
				logger.Error("on receive new view", "error", err)
			}
		}
	}
}

func (gns *genesis) onReceiveProposal(proposal *core.Block) error {
	logger.Debug("received proposal")
	if err := proposal.Validate(gns.resources.VldStore); err != nil {
		return err
	}
	if !proposal.IsGenesis() {
		return fmt.Errorf("not genesis block")
	}
	if !bytes.Equal(hashChainID(gns.chainID), proposal.ParentHash()) {
		return fmt.Errorf("different chain id genesis")
	}
	if !gns.isLeader(proposal.Proposer()) {
		return fmt.Errorf("proposer is not leader")
	}
	if len(proposal.Transactions()) != 0 {
		return fmt.Errorf("genesis block with txs")
	}
	gns.setB0(proposal)
	logger.Info("got genesis block, voting...")
	return gns.resources.MsgSvc.SendVote(proposal.Proposer(), proposal.Vote(gns.resources.Signer))
}

func (gns *genesis) onReceiveVote(vote *core.Vote) error {
	logger.Debug("received vote")
	if err := vote.Validate(gns.resources.VldStore); err != nil {
		return err
	}
	gns.mtxVote.Lock()
	defer gns.mtxVote.Unlock()

	gns.votes[vote.Voter().String()] = vote
	if len(gns.votes) < gns.resources.VldStore.ValidatorCount() {
		return nil
	}
	vlist := make([]*core.Vote, 0, len(gns.votes))
	for _, vote := range gns.votes {
		vlist = append(vlist, vote)
	}
	gns.setQ0(core.NewQuorumCert().Build(vlist))
	logger.Info("created qc, broadcasting...")
	gns.broadcastQC()
	return nil
}

func (gns *genesis) broadcastQC() {
	for {
		select {
		case <-gns.done:
			return
		default:
		}
		if err := gns.resources.MsgSvc.BroadcastNewView(gns.getQ0()); err != nil {
			logger.Error("broadcast proposal error", "error", err)
		}
		time.Sleep(time.Second)
	}
}

func (gns *genesis) onReceiveNewView(qc *core.QuorumCert) error {
	logger.Debug("received qc")
	gns.mtxNewView.Lock()
	defer gns.mtxNewView.Unlock()

	select {
	case <-gns.done:
		return nil
	default:
	}

	if err := qc.Validate(gns.resources.VldStore); err != nil {
		return err
	}
	b0 := gns.getB0()
	if b0 == nil {
		return fmt.Errorf("no received genesis block yet")
	}
	if !bytes.Equal(b0.Hash(), qc.BlockHash()) {
		return fmt.Errorf("invalid qc reference")
	}
	b0.SetQuorumCert(qc)
	gns.setQ0(qc)
	if !gns.isLeader(gns.resources.Signer.PublicKey()) {
		gns.resources.MsgSvc.SendNewView(b0.Proposer(), qc)
	}
	close(gns.done)
	return nil
}

func (gns *genesis) setB0(val *core.Block)      { gns.b0.Store(val) }
func (gns *genesis) setQ0(val *core.QuorumCert) { gns.q0.Store(val) }

func (gns *genesis) getB0() *core.Block {
	b0 := gns.b0.Load()
	if b0 == nil {
		return nil
	}
	return b0.(*core.Block)
}

func (gns *genesis) getQ0() *core.QuorumCert {
	q0 := gns.q0.Load()
	if q0 == nil {
		return nil
	}
	return q0.(*core.QuorumCert)
}
