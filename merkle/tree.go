// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package merkle

import (
	"crypto"
	"math/big"
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

// Update accepts new/modified tree leaves,
// recompute the corresponding nodes until root node.
func (tree *Tree) Update(leaves []*Node, newLeafCount *big.Int) *UpdateResult {
	res := &UpdateResult{newLeafCount, tree.calc.Height(newLeafCount), leaves}
	nodes := leaves
	for i := res.Height; i > 1; i-- {
		bpmap, nbmap := tree.groupNodesByBlock(nodes)
		blocks := tree.createBlocks(bpmap)
		parents := make([]*Node, 0, len(blocks)) // parent nodes

		for _, b := range blocks { // the body of the loop can run in parallel
			b.Load() // load blocks from store
			for _, n := range nbmap[b.parentPosition.String()] {
				b.SetNode(n) // set updated nodes in blocks
			}
			p := b.MakeParent()
			parents = append(parents, p)
			res.Nodes = append(res.Nodes, p)
		}
		nodes = parents
	}
	return res
}

func (tree *Tree) groupNodesByBlock(nodes []*Node) (map[string]*Position, map[string][]*Node) {
	nbps := tree.getBlockPositions(nodes)
	bpmap := nbps.UniqueMap()
	nbmap := make(map[string][]*Node, len(bpmap))
	for key := range bpmap {
		nbmap[key] = make([]*Node, 0)
	}
	for i, p := range nbps {
		nbmap[p.String()] = append(nbmap[p.String()], nodes[i])
	}
	return bpmap, nbmap
}

func (tree *Tree) getBlockPositions(nodes []*Node) Positions {
	positions := make(Positions, len(nodes))
	for i, n := range nodes {
		bIndex := tree.calc.BlockOfNode(n.Position.Index())
		positions[i] = NewPosition(n.Position.Level()+1, bIndex)
	}
	return positions
}

func (tree *Tree) createBlocks(pmap map[string]*Position) map[string]*Block {
	blocks := make(map[string]*Block, len(pmap))
	for _, p := range pmap {
		blocks[p.String()] = NewBlock(tree.hashFunc, tree.calc, tree.store, p)
	}
	return blocks
}
