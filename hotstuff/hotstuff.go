// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

// package hotstuff implements hotstuff core algorithm from https://arxiv.org/abs/1803.05069

package hotstuff

import (
	"context"
	"fmt"
)

// Hotstuff consensus engine
type Hotstuff struct {
	state
	driver Driver
}

// OnPropose is called to propose a new block
func (hs *Hotstuff) OnPropose(ctx context.Context) {
	bLeaf := hs.GetBLeaf()
	bNew := hs.driver.CreateLeaf(ctx, bLeaf, hs.GetQCHigh(), bLeaf.Height()+1)
	if bNew == nil {
		return
	}
	hs.setBLeaf(bNew)
	hs.startProposal(bNew)
	hs.driver.BroadcastProposal(bNew)
}

// OnReceiveVote is called when received a vote
func (hs *Hotstuff) OnReceiveVote(v Vote) {
	err := hs.addVote(v)
	if err != nil {
		return
	}
	if hs.GetVoteCount() >= hs.driver.MajorityCount() {
		qc := hs.driver.CreateQC(hs.GetVotes())
		hs.UpdateQCHigh(qc)
		hs.stopProposal()
	}
}

// OnReceiveProposal is called when a new proposal is received
func (hs *Hotstuff) OnReceiveProposal(bNew Block) {
	if hs.CanVote(bNew) {
		hs.driver.VoteBlock(bNew)
		hs.setVHeight(bNew.Height())
	}
	hs.Update(bNew)
}

// CanVote returns true if the hotstuff instance can vote the given block
func (hs *Hotstuff) CanVote(bNew Block) bool {
	if bNew.Height() <= hs.GetVHeight() {
		return false
	}
	return hs.CheckSafetyRule(bNew) || hs.CheckLivenessRule(bNew)
}

// CheckSafetyRule returns true if the given block extends from b_Lock
func (hs *Hotstuff) CheckSafetyRule(bNew Block) bool {
	bLock := hs.GetBLock()
	for b := bNew; CmpBlockHeight(b, bLock) != -1; b = b.Parent() {
		if bLock.Equal(b) {
			return true
		}
	}
	return false
}

// CheckLivenessRule returns true if the qc referenced block of the given block is higher than b_Lock
func (hs *Hotstuff) CheckLivenessRule(bNew Block) bool {
	return CmpBlockHeight(bNew.Justify().Block(), hs.GetBLock()) == 1
}

// Update perform three chain consensus phases
func (hs *Hotstuff) Update(bNew Block) {
	b, b1, b2 := GetJustifyBlocks(bNew)
	hs.UpdateQCHigh(bNew.Justify()) // precommit phase for b2
	if CmpBlockHeight(b1, hs.GetBLock()) == 1 {
		hs.setBLock(b1) // commit phase for b1
	} else {
		return
	}
	if IsThreeChain(b, b1, b2) {
		hs.onCommit(b) // decide phase for b
		hs.setBExec(b)
	}
}

func (hs *Hotstuff) onCommit(b Block) {
	if CmpBlockHeight(b, hs.GetBExec()) == 1 {
		// commit parent blocks recurrsively
		hs.onCommit(b.Parent())
		hs.driver.Execute(b)

	} else if !hs.GetBExec().Equal(b) {
		panic(fmt.Sprintf("hotstuff safety breached!!!\n%+v\n%+v\n", b, hs.GetBExec()))
	}
}

// UpdateQCHigh replaces qcHigh if the block of given qc is higher than the qcHigh block
func (hs *Hotstuff) UpdateQCHigh(qc QC) {
	if CmpBlockHeight(qc.Block(), hs.GetQCHigh().Block()) == 1 {
		hs.setQCHigh(qc)
		hs.setBLeaf(qc.Block())
	}
}
