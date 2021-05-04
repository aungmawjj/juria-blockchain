// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package storage

import "github.com/dgraph-io/badger"

// key prefixes for different data collections
const (
	_                  byte = iota
	keyBlock                // block by hash
	keyBlockHeight          // last block height
	keyBlockByHeight        // block hash by height
	keyBlockCommit          // block commit by hash
	keyTx                   // transaction by hash
	keyState                // state value by state key
	keyMerkleLeafIndex      // tree leaf index by state key
	keyMerkleHeight         // tree height
	keyMerkleLeafCount      // tree leaf count
	keyMerkleNode           // tree node value by position
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

func hasKey(db *badger.DB, key []byte) bool {
	err := db.View(func(txn *badger.Txn) error {
		_, err := txn.Get(key)
		return err
	})
	return err == nil
}
