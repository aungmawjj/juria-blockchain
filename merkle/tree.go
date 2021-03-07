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

func (t *Tree) blockCount(nodeCount *big.Int) *big.Int {
	// ceil(nodeCount / bfactor)
	count := big.NewInt(0)
	m := big.NewInt(0)
	count.DivMod(nodeCount, t.bfactor, m)
	if m.Cmp(big.NewInt(0)) == 1 {
		count.Add(count, big.NewInt(1))
	}
	return count
}

func (t Tree) firstNodeInBlock(blkIdx *big.Int) *big.Int {
	idx := big.NewInt(0)
	return idx.Mul(blkIdx, t.bfactor)
}

func (t *Tree) blockOfNode(nodeIdx *big.Int) *big.Int {
	// floor(nodeIdx / bfactor)
	idx := big.NewInt(0)
	return idx.Div(nodeIdx, t.bfactor)
}

func (t *Tree) nodeIndexInBlock(nodeIdx *big.Int) *big.Int {
	// mod(nodeIdx / bfactor)
	idx := big.NewInt(0)
	return idx.Mod(nodeIdx, t.bfactor)
}
