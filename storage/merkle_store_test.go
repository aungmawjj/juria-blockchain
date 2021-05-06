// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package storage

import (
	"math/big"
	"testing"

	"github.com/aungmawjj/juria-blockchain/merkle"
	"github.com/dgraph-io/badger/v3"
	"github.com/stretchr/testify/assert"
)

func createOnMemoryDB() *badger.DB {
	db, _ := badger.Open(badger.DefaultOptions("").WithInMemory(true))
	return db
}

func TestMerkleStore(t *testing.T) {
	assert := assert.New(t)

	db := createOnMemoryDB()
	ms := &MerkleStore{db}
	assert.Equal(uint8(0), ms.GetHeight())
	assert.Equal(big.NewInt(0), ms.GetLeafCount())

	assert.Nil(ms.GetNode(merkle.NewPosition(0, big.NewInt(0))))

	upd := &merkle.UpdateResult{
		LeafCount: big.NewInt(2),
		Height:    2,
		Leaves: []*merkle.Node{
			{Position: merkle.NewPosition(0, big.NewInt(0)), Data: []byte{1, 1}},
			{Position: merkle.NewPosition(0, big.NewInt(1)), Data: []byte{2, 2}},
		},
		Branches: []*merkle.Node{
			{Position: merkle.NewPosition(1, big.NewInt(0)), Data: []byte{3, 3}},
		},
	}

	updFuncs := ms.CommitUpdate(upd)
	updateDB(db, updFuncs)

	assert.Equal(upd.Height, ms.GetHeight())
	assert.Equal(upd.LeafCount, ms.GetLeafCount())
	assert.Equal([]byte{1, 1}, ms.GetNode(merkle.NewPosition(0, big.NewInt(0))))
	assert.Equal([]byte{2, 2}, ms.GetNode(merkle.NewPosition(0, big.NewInt(1))))
	assert.Equal([]byte{3, 3}, ms.GetNode(merkle.NewPosition(1, big.NewInt(0))))
}
