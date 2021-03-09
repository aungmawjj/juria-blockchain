// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package hotstuff

// Block type
type Block interface {
	Proposer() string
	Height() uint64
	Parent() Block
	Equal(blk Block) bool
	Justify() QC
}

// CmpBlockHeight compares two blocks by height
func CmpBlockHeight(b1, b2 Block) bool {
	if b1 == nil {
		return false
	}
	if b2 == nil {
		return true
	}
	return b1.Height() > b2.Height()
}

// QC type
type QC interface {
	Block() Block
}

// Vote type
type Vote interface {
	Block() Block
	Replica() string
}
