// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package merkle

import (
	"crypto"
	"math/big"
)

// Position of a node in the tree
type Position struct {
	Level uint8
	Index *big.Int
}

// Bytes binary value of a position
func (p *Position) Bytes() []byte {
	ib := p.Index.Bytes()
	b := make([]byte, 0, 1+len(ib))
	b = append(b, p.Level)
	b = append(b, ib...)
	return b
}

// TreeOptions type
type TreeOptions struct {
	BranchFactor uint8
	HashFunc     crypto.Hash
}

// Tree type
type Tree struct {
	bfactor *big.Int
}

// NewTree creates a new Merkle Tree
func NewTree(opts TreeOptions) *Tree {
	t := new(Tree)
	if opts.BranchFactor < 2 {
		opts.BranchFactor = 2
	}
	t.bfactor = big.NewInt(int64(opts.BranchFactor))
	return t
}
