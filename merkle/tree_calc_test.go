// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package merkle

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTreeCalc_Height(t *testing.T) {
	tests := []struct {
		name    string
		bfactor uint8
		nleaf   *big.Int
		want    uint8
	}{
		{"1 leaf", 4, big.NewInt(1), 1},
		{"2 leaf", 4, big.NewInt(2), 2},
		{"4 leaf", 4, big.NewInt(4), 2},
		{"6 leaf", 4, big.NewInt(6), 3},
		{"17 leaf, bf 8", 8, big.NewInt(17), 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewTreeCalc(tt.bfactor).Height(tt.nleaf)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTreeCalc_GroupCount(t *testing.T) {
	tests := []struct {
		name      string
		bfactor   uint8
		nodeCount *big.Int
		want      *big.Int
	}{
		{"one node", 4, big.NewInt(1), big.NewInt(1)},
		{"no remainder", 4, big.NewInt(8), big.NewInt(2)},
		{"has remainder", 4, big.NewInt(10), big.NewInt(3)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewTreeCalc(tt.bfactor).GroupCount(tt.nodeCount)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTreeCalc_FirstNodeInGroup(t *testing.T) {
	tests := []struct {
		name    string
		bfactor uint8
		blkIdx  *big.Int
		want    *big.Int
	}{
		{"block zero", 4, big.NewInt(0), big.NewInt(0)},
		{"block 1", 4, big.NewInt(1), big.NewInt(4)},
		{"block 2", 4, big.NewInt(2), big.NewInt(8)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewTreeCalc(tt.bfactor).FirstNodeOfGroup(tt.blkIdx)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTreeCalc_GroupOfNode(t *testing.T) {
	tests := []struct {
		name    string
		bfactor uint8
		nodeIdx *big.Int
		want    *big.Int
	}{
		{"node zero", 4, big.NewInt(0), big.NewInt(0)},
		{"first node in block 1", 4, big.NewInt(4), big.NewInt(1)},
		{"last node in block 1", 4, big.NewInt(7), big.NewInt(1)},
		{"first node in block 2", 4, big.NewInt(8), big.NewInt(2)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewTreeCalc(tt.bfactor).GroupOfNode(tt.nodeIdx)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTreeCalc_NodeIndexInGroup(t *testing.T) {
	tests := []struct {
		name    string
		bfactor uint8
		nodeIdx *big.Int
		want    int
	}{
		{"first node in block 0", 4, big.NewInt(0), 0},
		{"first node in block 1", 4, big.NewInt(4), 0},
		{"second node in block 1", 4, big.NewInt(5), 1},
		{"last node in block 2", 4, big.NewInt(11), 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewTreeCalc(tt.bfactor).NodeIndexInGroup(tt.nodeIdx)
			assert.Equal(t, tt.want, got)
		})
	}
}
