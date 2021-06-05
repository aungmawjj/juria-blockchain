// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package merkle

import (
	"bytes"
	"crypto"
	"math/big"
)

type Config struct {
	Hash            crypto.Hash
	BranchFactor    uint8
	ConcurrentLimit int
}

// Tree implements a merkle tree engine
type Tree struct {
	store  Store
	config Config
	calc   *TreeCalc
}

// NewTree creates a new Merkle Tree
func NewTree(store Store, config Config) *Tree {
	tree := new(Tree)
	tree.store = store
	tree.config = config
	if tree.config.BranchFactor < 2 {
		tree.config.BranchFactor = 2
	}
	if tree.config.ConcurrentLimit == 0 {
		tree.config.ConcurrentLimit = 20
	}
	tree.calc = NewTreeCalc(tree.config.BranchFactor)
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
	rowSize := newLeafCount

	for i := uint8(0); i < res.Height-1; i++ {
		nodes = tree.updateOneLevel(nodes, rowSize)
		res.Branches = append(res.Branches, nodes...)
		rowSize = tree.calc.GroupCount(rowSize)
	}
	if res.Height > 1 {
		res.Root = res.Branches[len(res.Branches)-1]
	} else {
		res.Root = res.Leaves[0]
	}
	return res
}

func (tree *Tree) updateOneLevel(nodes []*Node, rowSize *big.Int) []*Node {
	groups, gnodes := tree.groupNodesByParent(nodes)
	parents := make([]*Node, 0, len(groups))
	jobs, out := tree.spawnWorkers(rowSize, len(groups))
	defer close(jobs) // to stop workers
	for _, g := range groups {
		for _, n := range gnodes[g.parentPosition.String()] {
			g.SetNode(n) // set updated nodes in blocks
		}
		jobs <- g
	}
	for i := 0; i < len(groups); i++ {
		p := <-out
		parents = append(parents, p)
	}
	return parents
}

func (tree *Tree) spawnWorkers(rowSize *big.Int, nResult int) (chan *Group, chan *Node) {
	jobs := make(chan *Group, tree.config.ConcurrentLimit)
	out := make(chan *Node, nResult)
	for i := 0; i < tree.config.ConcurrentLimit; i++ {
		go tree.worker(rowSize, jobs, out)
	}
	return jobs, out
}

func (tree *Tree) worker(rowSize *big.Int, jobs <-chan *Group, out chan<- *Node) {
	for g := range jobs {
		g.Load(rowSize)
		out <- g.MakeParent()
	}
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
		groups[p.String()] = NewGroup(tree.config.Hash, tree.calc, tree.store, p)
	}
	return groups
}
