// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package storage

import (
	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/util"
	"github.com/dgraph-io/badger/v3"
)

type StateStore struct {
	db *badger.DB
}

func (ss *StateStore) GetState(codeAddr, key []byte) ([]byte, error) {
	return getValue(ss.db, util.ConcatBytes([]byte{colStateValueByKey}, codeAddr, key))
}

func (ss *StateStore) CommitUpdate(stateChanges []*core.StateChange) []updateFunc {
	ret := make([]updateFunc, 0, len(stateChanges))
	for _, sc := range stateChanges {
		ret = append(ret, ss.updateState(sc))
	}
	return ret
}

func (ss *StateStore) updateState(sc *core.StateChange) updateFunc {
	return func(txn *badger.Txn) error {
		key := util.ConcatBytes([]byte{colStateValueByKey}, sc.Key())
		if sc.Deleted() {
			return txn.Delete(key)
		}
		return txn.Set(key, sc.Value())
	}
}
