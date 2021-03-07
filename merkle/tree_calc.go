// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package merkle

import "math/big"

// TreeCalc calculates merkle tree properties
type TreeCalc struct {
	bfactor *big.Int
}

// Height gives the height of the tree based on number of leaves
func (t *TreeCalc) Height(nleaf *big.Int) uint8 {
	var h uint8 = 1
	nodeCount := nleaf
	for nodeCount.Cmp(big.NewInt(1)) == 1 { // node count > 1 (current level)
		h++
		nodeCount = t.BlockCount(nodeCount) // block count equals to node count of next level
	}
	return h
}

// BlockCount gives the number of blocks for a tree level with given node count
//
// e.g branch factor 5
//     [0           1] 		// 2 parent nodes
// [0 1 2 3 4] [5 6 7 _ _]	// 8 nodes becomes 2 blocks
//
func (t *TreeCalc) BlockCount(nodeCount *big.Int) *big.Int {
	// ceil(nodeCount / bfactor)
	count := big.NewInt(0)
	m := big.NewInt(0)
	count.DivMod(nodeCount, t.bfactor, m)
	if m.Cmp(big.NewInt(0)) == 1 {
		count.Add(count, big.NewInt(1))
	}
	return count
}

// FirstNodeOfBlock gives the index of first node of a block
//
// e.g branch factor 5
// [0 1 2 3 4] [5 6 7 _ _]	// first node of block 1 is 0 and of block 2 is 5
//
func (t TreeCalc) FirstNodeOfBlock(blkIdx *big.Int) *big.Int {
	idx := big.NewInt(0)
	return idx.Mul(blkIdx, t.bfactor)
}

// BlockOfNode gives the block index in which the node exist
//
// e.g branch factor 5
// [0 1 2 3 4] [5 6 7 _ _]	// block of node 1 is 0 and of node 7 is 2
//
func (t *TreeCalc) BlockOfNode(nodeIdx *big.Int) *big.Int {
	// floor(nodeIdx / bfactor)
	idx := big.NewInt(0)
	return idx.Div(nodeIdx, t.bfactor)
}

// NodeIndexInBlock gives the index of node in the corresponding block
//
// e.g branch factor 5
// [0 1 2 3 4] [5 6 7 _ _]	// indexInBlock of node 2 is 2 and of node 6 is 1
//
func (t *TreeCalc) NodeIndexInBlock(nodeIdx *big.Int) *big.Int {
	// mod(nodeIdx / bfactor)
	idx := big.NewInt(0)
	return idx.Mod(nodeIdx, t.bfactor)
}
