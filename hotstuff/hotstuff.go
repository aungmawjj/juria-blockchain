// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

// package hotstuff implements hotstuff core algorithm from https://arxiv.org/abs/1803.05069

package hotstuff

import "context"

// Hotstuff consensus engine
type Hotstuff struct {
	state
	driver Driver
}

// Init initializes hotstuff
func (hs *Hotstuff) Init(b0 Block, q0 QC) {
	hs.state.init(b0, q0)
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
	hs.driver.SendProposal(bNew)
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
	}
}

// UpdateQCHigh replaces qcHigh if the block of given qc is higher than the qcHigh block
func (hs *Hotstuff) UpdateQCHigh(qc QC) {
	if CmpBlockHeight(qc.Block(), hs.GetQCHigh().Block()) == 1 {
		hs.setQCHigh(qc)
		hs.setBLeaf(qc.Block())
	}
}
