// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package hotstuff

// Block type
type Block interface {
	Height() uint64
	Parent() Block
	Equal(blk Block) bool
	Justify() QC
}

// QC type
type QC interface {
	Block() Block
}

// Vote type
type Vote interface {
	Block() Block
	Voter() string
}

// Driver godoc
type Driver interface {
	MajorityCount() int
	CreateLeaf(parent Block, qc QC, height uint64) Block
	CreateQC(votes []Vote) QC
	BroadcastProposal(blk Block)
	VoteBlock(blk Block)
	Commit(blk Block)
}

// CmpBlockHeight compares two blocks by height
func CmpBlockHeight(b1, b2 Block) int {
	if b1 == nil && b2 == nil {
		return 0
	}
	if b1 == nil {
		return -1
	}
	if b2 == nil {
		return 1
	}
	if b1.Height() == b2.Height() {
		return 0
	}
	if b1.Height() > b2.Height() {
		return 1
	}
	return -1
}

// GetJustifyBlocks returns justify referenced blocks
func GetJustifyBlocks(bNew Block) (b, b1, b2 Block) {
	if b2 = bNew.Justify().Block(); b2 == nil {
		return b, b1, b2
	}
	if b1 = b2.Justify().Block(); b1 == nil {
		return b, b1, b2
	}
	b = b1.Justify().Block()
	return b, b1, b2
}

// IsThreeChain checks whether the blocks satisfy three chain rule
func IsThreeChain(b, b1, b2 Block) bool {
	if b == nil || b1 == nil || b2 == nil {
		return false
	}
	return b1.Equal(b2.Parent()) && b.Equal(b1.Parent())
}
