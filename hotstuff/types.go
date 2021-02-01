// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package hotstuff

// Block godoc
type Block interface {
	Proposer() string
	Height() uint64
	Parent() Block
	Equal(blk Block) bool
	Justify() QC
}

// QC godoc
type QC interface {
	Block() Block
}

// Vote godoc
type Vote interface {
	Block() Block
	Replica() string
}
