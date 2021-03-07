// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package merkle

import (
	"crypto"
)

// TreeOptions type
type TreeOptions struct {
	BranchFactor uint8
	HashFunc     crypto.Hash
}

// Tree type
type Tree struct {
	bfactor uint8
}

// NewTree creates a new Merkle Tree
func NewTree(opts TreeOptions) *Tree {
	t := new(Tree)
	if opts.BranchFactor < 2 {
		t.bfactor = 2
	} else {
		t.bfactor = opts.BranchFactor
	}
	return t
}
