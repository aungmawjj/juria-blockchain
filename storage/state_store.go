// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package storage

import (
	"bytes"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/dgraph-io/badger/v3"
)

type StateStore struct {
	db *badger.DB
}

func (ss *StateStore) GetState(key []byte) ([]byte, error) {
	return getValue(ss.db, concatBytes([]byte{colStateValueByKey}, key))
}

func (ss *StateStore) getMerkleIndex(key []byte) ([]byte, error) {
	return getValue(ss.db, concatBytes([]byte{colMerkleIndexByStateKey}, key))
}

func (ss *StateStore) updateState(sc *core.StateChange) updateFunc {
	return func(txn *badger.Txn) error {
		err := txn.Set(
			concatBytes([]byte{colStateValueByKey}, sc.Key()), sc.Value(),
		)
		if err != nil {
			return err
		}
		if bytes.Equal(sc.TreeIndex(), sc.PrevTreeIndex()) {
			return nil
		}
		return txn.Set(
			concatBytes([]byte{colMerkleIndexByStateKey}, sc.Key()), sc.TreeIndex(),
		)
	}
}
