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

// Update ...
func (tree *Tree) Update(leaves []*Node, newLeafCount *big.Int) *UpdateResult {
	res := &UpdateResult{newLeafCount, tree.calc.Height(newLeafCount), leaves}

	nodes := leaves
	for i := 1; i < int(res.Height); i++ {

		nbPositions := tree.getBlockPositions(nodes)
		bPositions := tree.mergePositions(nbPositions)
		nodesByBlock := tree.groupNodesByBlock(nodes, nbPositions, bPositions)

		blocks := tree.createBlocks(bPositions)
		parents := make([]*Node, 0, len(blocks)) // parent nodes

		for _, b := range blocks { // the body of the loop can be run in parallel
			b.Load() // load blocks from store
			for _, n := range nodesByBlock[b.parentPosition.String()] {
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

func (tree *Tree) groupNodesByBlock(
	nodes []*Node, nbPositions []*Position, bPositions map[string]*Position,
) map[string][]*Node {
	nb := make(map[string][]*Node, len(bPositions))
	for key := range bPositions {
		nb[key] = make([]*Node, 0)
	}
	for i, p := range nbPositions {
		nb[p.String()] = append(nb[p.String()], nodes[i])
	}
	return nb
}

func (tree *Tree) getBlockPositions(nodes []*Node) []*Position {
	positions := make([]*Position, len(nodes))
	for i, n := range nodes {
		bIndex := tree.calc.BlockOfNode(n.Position.Index())
		positions[i] = NewPosition(n.Position.Level()+1, bIndex)
	}
	return positions
}

func (tree *Tree) mergePositions(positions []*Position) map[string]*Position {
	pmap := make(map[string]*Position)
	for _, p := range positions {
		if _, found := pmap[p.String()]; !found {
			pmap[p.String()] = p
		}
	}
	return pmap
}

func (tree *Tree) createBlocks(pmap map[string]*Position) map[string]*Block {
	blocks := make(map[string]*Block, len(pmap))
	for _, p := range pmap {
		blocks[p.String()] = NewBlock(tree.hashFunc, tree.calc, tree.store, p)
	}
	return blocks
}
