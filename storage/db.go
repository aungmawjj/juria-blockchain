// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package storage

import (
	"bytes"

	"github.com/dgraph-io/badger/v3"
)

// data collection prefixes for different data collections
const (
	_                        byte = iota
	colBlockByHash                // block by hash
	colBlockHashByHeight          // block hash by height
	colBlockHeight                // last block height
	colBlockCommitByHash          // block commit by block hash
	colTxByHash                   // tx by hash
	colTxCommitByHash             // tx commit info by tx hash
	colStateValueByKey            // state value by state key
	colMerkleIndexByStateKey      // tree leaf index by state key
	colMerkleTreeHeight           // tree height
	colMerkleLeafCount            // tree leaf count
	colMerkleNodeByPosition       // tree node value by position
)

type setter interface {
	Set(key, value []byte) error
}

type updateFunc func(setter setter) error

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

func updateDB(db *badger.DB, fns []updateFunc) error {
	return db.Update(func(txn *badger.Txn) error {
		for _, fn := range fns {
			if err := fn(txn); err != nil {
				return err
			}
		}
		return nil
	})
}

func concatBytes(srcs ...[]byte) []byte {
	buf := bytes.NewBuffer(nil)
	for _, src := range srcs {
		buf.Grow(len(src))
	}
	for _, src := range srcs {
		buf.Write(src)
	}
	return buf.Bytes()
}
