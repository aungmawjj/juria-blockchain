// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package storage

import (
	"bytes"

	"github.com/dgraph-io/badger/v3"
)

// data collection prefixes for different data collections
const (
	colBlockByHash           byte = iota + 1 // block by hash
	colBlockHashByHeight                     // block hash by height
	colBlockHeight                           // last block height
	colLastQC                                // qc for last commited block to be used on restart
	colBlockCommitByHash                     // block commit by block hash
	colTxCount                               // total commited tx count
	colTxByHash                              // tx by hash
	colTxCommitByHash                        // tx commit info by tx hash
	colStateValueByKey                       // state value by state key
	colMerkleIndexByStateKey                 // tree leaf index by state key
	colMerkleTreeHeight                      // tree height
	colMerkleLeafCount                       // tree leaf count
	colMerkleNodeByPosition                  // tree node value by position
)

func NewDB(path string) (*badger.DB, error) {
	return badger.Open(badger.DefaultOptions(path))
}

type setter interface {
	Set(key, value []byte) error
}

type updateFunc func(setter setter) error

type getter interface {
	Get(key []byte) ([]byte, error)
	HasKey(key []byte) bool
}

type badgerGetter struct {
	db *badger.DB
}

func (bg *badgerGetter) Get(key []byte) ([]byte, error) {
	var val []byte
	err := bg.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err == nil {
			val, err = item.ValueCopy(nil)
		}
		return err
	})
	return val, err
}

func (bg *badgerGetter) HasKey(key []byte) bool {
	err := bg.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get(key)
		return err
	})
	return err == nil
}

func updateBadgerDB(db *badger.DB, fns []updateFunc) error {
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
	size := 0
	for _, src := range srcs {
		size += len(src)
	}
	buf.Grow(size)
	for _, src := range srcs {
		buf.Write(src)
	}
	return buf.Bytes()
}
