// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package storage

import "github.com/dgraph-io/badger"

// key prefixes for different data collections
const (
	_ byte = iota
	keyBlock
	keyBlockByHeight
	keyTx
	keyTxBySender
	keyTxByCodeAddr
	keyState
	keyMerkleLeafIndex
	keyMerkleHeight
	keyMerkleLeafCount
	keyMerkleNode
)

type updateFunc func(txn *badger.Txn) error

func getValue(db *badger.DB, key []byte) ([]byte, error) {
	var val []byte
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err == nil {
			val, err = item.ValueCopy(nil)
		}
		return err
	})
	return val, err
}
