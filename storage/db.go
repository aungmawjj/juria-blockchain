// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package storage

import "github.com/dgraph-io/badger"

// data collection prefixes for different data collections
const (
	_                        byte = iota
	colBlockByHash                // block by hash
	colBlockByHeight              // block hash by height
	colBlockHeight                // last block height
	colBlockCommitByHash          // block commit by hash
	colTxByHash                   // transaction by hash
	colStateValueByKey            // state value by state key
	colMerkleIndexByStateKey      // tree leaf index by state key
	colMerkleTreeHeight           // tree height
	colMerkleLeafCount            // tree leaf count
	colMerkleNodeByPosition       // tree node value by position
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
