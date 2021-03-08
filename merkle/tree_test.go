// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package merkle

import (
	"crypto"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTree(t *testing.T) {
	tests := []struct {
		name string
		opts TreeOptions
		want uint8
	}{
		{"branch factor < 2", TreeOptions{1, crypto.SHA1}, 2},
		{"branch factor >= 2", TreeOptions{4, crypto.SHA1}, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := NewTree(nil, tt.opts)
			assert.Equal(t, tt.want, tree.bfactor)
		})
	}
}

func TestTree_Update(t *testing.T) {
	store := NewMapStore()
	var bfactor uint8 = 3
	tree := NewTree(store, TreeOptions{bfactor, crypto.SHA1})

	leaves := make([]*Node, 7)
	for i := range leaves {
		leaves[i] = &Node{NewPosition(0, big.NewInt(int64(i))), []byte{uint8(i)}}
	}

	res := tree.Update(leaves, big.NewInt(7))
	store.CommitUpdate(res)

	n10 := sha1Sum([]byte{0, 1, 2}) // level 0, index 1
	n11 := sha1Sum([]byte{3, 4, 5})
	n12 := sha1Sum([]byte{6})
	n20 := sha1Sum(append(n10, append(n11, n12...)...))

	assert := assert.New(t)

	assert.Equal(11, len(store.nodes))
	assert.Equal(n10, store.GetNode(NewPosition(1, big.NewInt(0))))
	assert.Equal(n11, store.GetNode(NewPosition(1, big.NewInt(1))))
	assert.Equal(n20, store.GetNode(NewPosition(2, big.NewInt(0))))

	upd := []*Node{
		{NewPosition(0, big.NewInt(2)), []byte{1}},
		{NewPosition(0, big.NewInt(5)), []byte{1}},
		{NewPosition(0, big.NewInt(7)), []byte{1}},
		{NewPosition(0, big.NewInt(8)), []byte{1}},
		{NewPosition(0, big.NewInt(9)), []byte{1}},
	}
	res = tree.Update(upd, big.NewInt(10))
	store.CommitUpdate(res)

	nn10 := sha1Sum([]byte{0, 1, 1})
	nn11 := sha1Sum([]byte{3, 4, 1})
	nn12 := sha1Sum([]byte{6, 1, 1})
	nn13 := sha1Sum([]byte{1})
	nn20 := sha1Sum(append(nn10, append(nn11, nn12...)...))
	nn21 := sha1Sum(nn13)
	nn30 := sha1Sum(append(nn20, nn21...))

	assert.Equal(17, len(store.nodes))
	assert.Equal(nn10, store.GetNode(NewPosition(1, big.NewInt(0))))
	assert.Equal(nn11, store.GetNode(NewPosition(1, big.NewInt(1))))
	assert.Equal(nn12, store.GetNode(NewPosition(1, big.NewInt(2))))
	assert.Equal(nn13, store.GetNode(NewPosition(1, big.NewInt(3))))
	assert.Equal(nn20, store.GetNode(NewPosition(2, big.NewInt(0))))
	assert.Equal(nn21, store.GetNode(NewPosition(2, big.NewInt(1))))
	assert.Equal(nn30, store.GetNode(NewPosition(3, big.NewInt(0))))
}
