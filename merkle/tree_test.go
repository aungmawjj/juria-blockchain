package merkle

import (
	"crypto"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPosition_Bytes(t *testing.T) {
	tests := []struct {
		name  string
		level uint8
		index *big.Int
		want  []byte
	}{
		{"level 0, index 0", 0, big.NewInt(0), []byte{0}},
		{"index 0", 1, big.NewInt(0), []byte{1}},
		{"index max 8 bit", 1, big.NewInt(255), []byte{1, 255}},
		{"index first 16 bit", 1, big.NewInt(256), []byte{1, 1, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Position{tt.level, tt.index}
			assert.EqualValues(t, tt.want, p.Bytes())
		})
	}
}

func TestNewTree(t *testing.T) {
	tests := []struct {
		name string
		opts TreeOptions
		want *big.Int
	}{
		{"branch factor < 2", TreeOptions{1, crypto.SHA1}, big.NewInt(2)},
		{"branch factor >= 2", TreeOptions{4, crypto.SHA1}, big.NewInt(4)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := NewTree(tt.opts)
			assert.Equal(t, 0, tt.want.Cmp(tree.bfactor))
		})
	}
}

func TestTree_blockCount(t *testing.T) {
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
			tree := NewTree(TreeOptions{tt.bfactor, crypto.SHA1})
			got := tree.blockCount(tt.nodeCount)
			assert.Equal(t, 0, tt.want.Cmp(got))
		})
	}
}

func TestTree_firstNodeInBlock(t *testing.T) {
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
			tree := NewTree(TreeOptions{tt.bfactor, crypto.SHA1})
			got := tree.firstNodeInBlock(tt.blkIdx)
			assert.Equal(t, 0, tt.want.Cmp(got))
		})
	}
}

func TestTree_blockOfNode(t *testing.T) {
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
			tree := NewTree(TreeOptions{tt.bfactor, crypto.SHA1})
			got := tree.blockOfNode(tt.nodeIdx)
			assert.Equal(t, 0, tt.want.Cmp(got))
		})
	}
}

func TestTree_nodeIndexInBlock(t *testing.T) {
	tests := []struct {
		name    string
		bfactor uint8
		nodeIdx *big.Int
		want    *big.Int
	}{
		{"first node in block 0", 4, big.NewInt(0), big.NewInt(0)},
		{"first node in block 1", 4, big.NewInt(4), big.NewInt(0)},
		{"second node in block 1", 4, big.NewInt(5), big.NewInt(1)},
		{"last node in block 2", 4, big.NewInt(11), big.NewInt(3)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := NewTree(TreeOptions{tt.bfactor, crypto.SHA1})
			got := tree.nodeIndexInBlock(tt.nodeIdx)
			assert.Equal(t, 0, tt.want.Cmp(got))
		})
	}
}
