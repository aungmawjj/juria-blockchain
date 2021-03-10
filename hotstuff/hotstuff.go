// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

// package hotstuff implements hotstuff core algorithm from https://arxiv.org/abs/1803.05069

package hotstuff

// Hotstuff consensus engine
type Hotstuff struct {
	state
	driver Driver
}

// Init initializes hotstuff
func (hs *Hotstuff) Init(b0 Block, q0 QC) {
	hs.state.init(b0, q0)
}

// OnNextSyncView changes view
func (hs *Hotstuff) OnNextSyncView() {
	hs.driver.SendNewView(hs.GetQCHigh())
}

// OnReceiveNewView is invoked when a new view is received from a peer
func (hs *Hotstuff) OnReceiveNewView(qc QC) {
	hs.updateQCHigh(qc)
}

func (hs *Hotstuff) updateQCHigh(qc QC) {
	if CmpBlockHeight(qc.Block(), hs.GetQCHigh().Block()) {
		hs.setQCHigh(qc)
		hs.setBLeaf(qc.Block())
	}
}
