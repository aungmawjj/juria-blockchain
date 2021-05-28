// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package merkle

import (
	"bytes"
	"crypto"
	"math/big"
)

// Tree implements a merkle tree engine
type Tree struct {
	store    Store
	bfactor  uint8
	hashFunc crypto.Hash
	calc     *TreeCalc
}

// NewTree creates a new Merkle Tree
func NewTree(store Store, h crypto.Hash, bfactor uint8) *Tree {
	tree := new(Tree)
	tree.store = store
	tree.bfactor = bfactor
	if tree.bfactor < 2 {
		tree.bfactor = 2
	}
	tree.hashFunc = h
	tree.calc = NewTreeCalc(tree.bfactor)
	return tree
}

// Root returns the root node of the tree
func (tree *Tree) Root() *Node {
	p := NewPosition(tree.store.GetHeight()-1, big.NewInt(0))
	if data := tree.store.GetNode(p); data != nil {
		return &Node{p, data}
	}
	return nil
}

// Update accepts new/modified tree leaves,
// recompute the corresponding nodes until root node.
func (tree *Tree) Update(leaves []*Node, newLeafCount *big.Int) *UpdateResult {
	res := &UpdateResult{
		LeafCount: newLeafCount,
		Height:    tree.calc.Height(newLeafCount),
		Leaves:    leaves,
		Branches:  make([]*Node, 0),
	}
	nodes := leaves
	rowNodeCount := newLeafCount

	for i := res.Height; i > 1; i-- {
		groups, gnodes := tree.groupNodesByParent(nodes)
		parents := make([]*Node, 0, len(groups))
		for _, g := range groups { // the body of the loop can run in parallel
			g.Load(rowNodeCount)
			for _, n := range gnodes[g.parentPosition.String()] {
				g.SetNode(n) // set updated nodes in blocks
			}
			p := g.MakeParent()
			parents = append(parents, p)
			res.Branches = append(res.Branches, p)
		}
		nodes = parents
		rowNodeCount = tree.calc.GroupCount(rowNodeCount)
	}
	if res.Height > 1 {
		res.Root = res.Branches[len(res.Branches)-1]
	} else {
		res.Root = res.Leaves[0]
	}
	return res
}

// Verify verifies leaves with the current root-node.
func (tree *Tree) Verify(leaves []*Node) bool {
	root := tree.Root()
	if root == nil {
		return false
	}
	if len(leaves) == 0 {
		return false
	}
	leafCount := tree.store.GetLeafCount()
	if leafCount.Cmp(big.NewInt(0)) == 0 {
		return false
	}
	for _, n := range leaves {
		if n.Position.Level() != 0 {
			return false
		}
		if leafCount.Cmp(n.Position.Index()) != 1 { // leaf count must be larger than leaf index
			return false
		}
	}
	res := tree.Update(leaves, leafCount)
	return bytes.Equal(root.Data, res.Root.Data)
}

func (tree *Tree) groupNodesByParent(nodes []*Node) (map[string]*Group, map[string][]*Node) {
	ngmap := tree.getGroupPositions(nodes)
	bpos := ngmap.UniqueMap()
	groups := tree.makeGroups(bpos)
	gnodes := make(map[string][]*Node, len(bpos))
	for key := range bpos {
		gnodes[key] = make([]*Node, 0)
	}
	for i, p := range ngmap {
		gnodes[p.String()] = append(gnodes[p.String()], nodes[i])
	}
	return groups, gnodes
}

func (tree *Tree) getGroupPositions(nodes []*Node) Positions {
	positions := make(Positions, len(nodes))
	for i, n := range nodes {
		bIndex := tree.calc.GroupOfNode(n.Position.Index())
		positions[i] = NewPosition(n.Position.Level()+1, bIndex)
	}
	return positions
}

func (tree *Tree) makeGroups(pmap map[string]*Position) map[string]*Group {
	groups := make(map[string]*Group, len(pmap))
	for _, p := range pmap {
		groups[p.String()] = NewGroup(tree.hashFunc, tree.calc, tree.store, p)
	}
	return groups
}
