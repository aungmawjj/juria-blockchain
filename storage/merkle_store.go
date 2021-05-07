// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package storage

import (
	"math/big"

	"github.com/aungmawjj/juria-blockchain/merkle"
	"github.com/dgraph-io/badger/v3"
)

type MerkleStore struct {
	db *badger.DB
}

var _ merkle.Store = (*MerkleStore)(nil)

func (ms *MerkleStore) GetLeafCount() *big.Int {
	count := big.NewInt(0)
	val, err := getValue(ms.db, []byte{colMerkleLeafCount})
	if err == nil {
		count.SetBytes(val)
	}
	return count
}

func (ms *MerkleStore) GetHeight() uint8 {
	var height uint8
	val, _ := getValue(ms.db, []byte{colMerkleTreeHeight})
	if len(val) > 0 {
		height = val[0]
	}
	return height
}

func (ms *MerkleStore) GetNode(p *merkle.Position) []byte {
	val, _ := getValue(
		ms.db, concatBytes([]byte{colMerkleNodeByPosition}, p.Bytes()),
	)
	return val
}

func (ms *MerkleStore) commitUpdate(upd *merkle.UpdateResult) []updateFunc {
	ret := make([]updateFunc, 0)
	for _, n := range upd.Leaves {
		ret = append(ret, ms.storeNode(n))
	}
	for _, n := range upd.Branches {
		ret = append(ret, ms.storeNode(n))
	}
	ret = append(ret, ms.storeCount(upd))
	ret = append(ret, ms.storeTreeHeight(upd))
	return ret
}

func (ms *MerkleStore) storeCount(upd *merkle.UpdateResult) updateFunc {
	return func(txn *badger.Txn) error {
		return txn.Set([]byte{colMerkleLeafCount}, upd.LeafCount.Bytes())
	}
}

func (ms *MerkleStore) storeTreeHeight(upd *merkle.UpdateResult) updateFunc {
	return func(txn *badger.Txn) error {
		return txn.Set([]byte{colMerkleTreeHeight}, []byte{upd.Height})
	}
}

func (ms *MerkleStore) storeNode(n *merkle.Node) updateFunc {
	return func(txn *badger.Txn) error {
		return txn.Set(
			concatBytes([]byte{colMerkleNodeByPosition}, n.Position.Bytes()), n.Data,
		)
	}
}
