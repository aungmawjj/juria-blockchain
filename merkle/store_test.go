// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package merkle

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapStore(t *testing.T) {
	assert := assert.New(t)

	ms := NewMapStore()
	assert.Equal(uint8(0), ms.GetHeight())
	assert.Equal(big.NewInt(0), ms.GetLeafCount())

	leafZero := ms.GetNode(NewPosition(0, big.NewInt(0)))
	assert.Nil(leafZero)

	upd := &UpdateResult{
		LeafCount: big.NewInt(2),
		Height:    2,
		Nodes: []*Node{
			{NewPosition(0, big.NewInt(0)), []byte{1, 1}},
			{NewPosition(0, big.NewInt(1)), []byte{2, 2}},
			{NewPosition(1, big.NewInt(0)), []byte{3, 3}},
		},
	}

	ms.CommitUpdate(upd)

	assert.Equal(upd.Height, ms.GetHeight())
	assert.Equal(upd.LeafCount, ms.GetLeafCount())
	assert.Equal([]byte{1, 1}, ms.GetNode(NewPosition(0, big.NewInt(0))))
	assert.Equal([]byte{2, 2}, ms.GetNode(NewPosition(0, big.NewInt(1))))
	assert.Equal([]byte{3, 3}, ms.GetNode(NewPosition(1, big.NewInt(0))))
}
