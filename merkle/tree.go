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
	store    Store
	bfactor  uint8
	hashFunc crypto.Hash
	calc     *TreeCalc
}

// NewTree creates a new Merkle Tree
func NewTree(store Store, opts TreeOptions) *Tree {
	tree := new(Tree)
	tree.store = store
	if opts.BranchFactor < 2 {
		tree.bfactor = 2
	} else {
		tree.bfactor = opts.BranchFactor
	}
	tree.hashFunc = opts.HashFunc
	tree.calc = NewTreeCalc(tree.bfactor)
	return tree
}
