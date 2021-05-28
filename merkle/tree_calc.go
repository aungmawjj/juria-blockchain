// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package merkle

import "math/big"

// TreeCalc calculates merkle tree properties
type TreeCalc struct {
	bfactorRaw uint8
	bfactor    *big.Int
}

// NewTreeCalc creates a new NewTreeCalc
func NewTreeCalc(bfactor uint8) *TreeCalc {
	return &TreeCalc{bfactor, big.NewInt(int64(bfactor))}
}

// BranchFactor returns uint8 branch factor
func (tc *TreeCalc) BranchFactor() uint8 {
	return tc.bfactorRaw
}

// Height gives the height of the tree based on number of leaves
func (tc *TreeCalc) Height(nleaf *big.Int) uint8 {
	var h uint8 = 1
	nodeCount := nleaf
	for nodeCount.Cmp(big.NewInt(1)) == 1 { // node count > 1 (current level)
		h++
		nodeCount = tc.GroupCount(nodeCount) // block count equals to node count of next level
	}
	return h
}

// GroupCount gives the number of groups for a tree level with given node count
//
// e.g branch factor 5
//     [0           1] 		// 2 parent nodes
// [0 1 2 3 4] [5 6 7 _ _]	// 8 nodes becomes 2 blocks
//
func (tc *TreeCalc) GroupCount(nodeCount *big.Int) *big.Int {
	// ceil(nodeCount / bfactor)
	count := big.NewInt(0)
	m := big.NewInt(0)
	count.DivMod(nodeCount, tc.bfactor, m)
	if m.Cmp(big.NewInt(0)) == 1 {
		count.Add(count, big.NewInt(1))
	}
	return count
}

// FirstNodeOfGroup gives the index of first node of a group
//
// e.g branch factor 5
// [0 1 2 3 4] [5 6 7 _ _]	// first node of block 1 is 0 and of block 2 is 5
//
func (tc TreeCalc) FirstNodeOfGroup(blkIdx *big.Int) *big.Int {
	idx := big.NewInt(0)
	return idx.Mul(blkIdx, tc.bfactor)
}

// GroupOfNode gives the group index in which the node exist
//
// e.g branch factor 5
// [0 1 2 3 4] [5 6 7 _ _]	// group of node 1 is 0 and of node 7 is 2
//
func (tc *TreeCalc) GroupOfNode(nodeIdx *big.Int) *big.Int {
	// floor(nodeIdx / bfactor)
	idx := big.NewInt(0)
	return idx.Div(nodeIdx, tc.bfactor)
}

// NodeIndexInGroup gives the index of node in the corresponding group
//
// e.g branch factor 5
// [0 1 2 3 4] [5 6 7 _ _]	// indexInGroup of node 2 is 2 and of node 6 is 1
//
func (tc *TreeCalc) NodeIndexInGroup(nodeIdx *big.Int) int {
	// mod(nodeIdx / bfactor)
	idx := big.NewInt(0)
	return int(idx.Mod(nodeIdx, tc.bfactor).Int64())
}
