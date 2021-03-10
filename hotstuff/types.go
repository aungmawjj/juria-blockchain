// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package hotstuff

import "context"

// Block type
type Block interface {
	Proposer() string
	Height() uint64
	Parent() Block
	Equal(blk Block) bool
	Justify() QC
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

// QC type
type QC interface {
	Block() Block
}

// Vote type
type Vote interface {
	Block() Block
	Replica() string
}

// Driver godoc
type Driver interface {
	CreateLeaf(ctx context.Context, parent Block, qc QC, height uint64) Block
	CreateQC(votes []Vote) QC
	SendProposal(blk Block)
	SendVote(blk Block)
	Execute(blk Block)
	MajorityCount() int
}
