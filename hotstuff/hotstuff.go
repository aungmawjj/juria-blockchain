// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package hotstuff

// Hotstuff consensus engine
type Hotstuff struct {
	state
}

// Init initializes hotstuff
func (hs *Hotstuff) Init(b0 Block, q0 QC) {
	hs.state.init(b0, q0)
}

// OnReceiveNewView should be invoked when a new view is received from a peer
func (hs *Hotstuff) OnReceiveNewView(qc QC) {
	hs.updateQCHigh(qc)
}

func (hs *Hotstuff) updateQCHigh(qc QC) {
	if CmpBlockHeight(qc.Block(), hs.GetQCHigh().Block()) {
		hs.setQCHigh(qc)
		hs.setBLeaf(qc.Block())
	}
}
