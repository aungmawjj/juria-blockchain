// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package storage

import (
	"math/big"

	"github.com/aungmawjj/juria-blockchain/merkle"
	"github.com/aungmawjj/juria-blockchain/util"
	"github.com/dgraph-io/badger"
)

type MerkleStore struct {
	db *badger.DB
}

var _ merkle.Store = (*MerkleStore)(nil)

func (ms *MerkleStore) GetLeafCount() *big.Int {
	count := big.NewInt(0)
	val, err := getValue(ms.db, []byte{keyMerkleLeafCount})
	if err == nil {
		count.SetBytes(val)
	}
	return count
}

func (ms *MerkleStore) GetHeight() uint8 {
	var height uint8
	val, _ := getValue(ms.db, []byte{keyMerkleHeight})
	if len(val) > 0 {
		height = val[0]
	}
	return height
}

func (ms *MerkleStore) GetNode(p *merkle.Position) []byte {
	val, _ := getValue(ms.db, ms.nodeKey(p))
	return val
}

func (ms *MerkleStore) CommitUpdate(upd *merkle.UpdateResult) []updateFunc {
	ret := ms.setNodes(upd)
	ret = append(ret, ms.setLeafCount(upd))
	ret = append(ret, ms.setHeight(upd))
	return ret
}

func (ms *MerkleStore) setNodes(upd *merkle.UpdateResult) []updateFunc {
	ret := make([]updateFunc, 0, len(upd.Branches)+len(upd.Leaves))
	for _, n := range upd.Leaves {
		ret = append(ret, ms.setNode(n))
	}
	for _, n := range upd.Branches {
		ret = append(ret, ms.setNode(n))
	}
	return ret
}

func (ms *MerkleStore) setLeafCount(upd *merkle.UpdateResult) updateFunc {
	return func(txn *badger.Txn) error {
		return txn.Set([]byte{keyMerkleLeafCount}, upd.LeafCount.Bytes())
	}
}

func (ms *MerkleStore) setHeight(upd *merkle.UpdateResult) updateFunc {
	return func(txn *badger.Txn) error {
		return txn.Set([]byte{keyMerkleHeight}, []byte{upd.Height})
	}
}

func (ms *MerkleStore) setNode(n *merkle.Node) updateFunc {
	return func(txn *badger.Txn) error {
		return txn.Set(ms.nodeKey(n.Position), n.Data)
	}
}

func (ms *MerkleStore) nodeKey(p *merkle.Position) []byte {
	return util.ConcatBytes([]byte{keyMerkleNode}, p.Bytes())
}
