// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package storage

import (
	"math/big"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/merkle"
	"github.com/dgraph-io/badger/v3"
	"golang.org/x/crypto/sha3"
)

type stateStore struct {
	db *badger.DB
}

func (ss *stateStore) loadPrevValues(scList []*core.StateChange) error {
	for _, sc := range scList {
		val, err := ss.getState(sc.Key())
		if err != nil {
			return err
		}
		sc.SetPrevValue(val)
	}
	return nil
}

func (ss *stateStore) loadPrevTreeIndexes(scList []*core.StateChange) error {
	for _, sc := range scList {
		val, err := ss.getMerkleIndex(sc.Key())
		if err != nil {
			return err
		}
		sc.SetPrevTreeIndex(val)
	}
	return nil
}

func (ss *stateStore) setNewTreeIndexes(leafCount *big.Int, scList []*core.StateChange) *big.Int {
	lc := big.NewInt(0).Set(leafCount)
	for _, sc := range scList {
		if sc.PrevTreeIndex() == nil {
			sc.SetTreeIndex(lc.Bytes())
			lc.Add(lc, big.NewInt(1))
		} else {
			sc.SetTreeIndex(sc.PrevTreeIndex())
		}
	}
	return lc
}

func (ss *stateStore) computeUpdatedTreeNodes(scList []*core.StateChange) []*merkle.Node {
	nodes := make([]*merkle.Node, len(scList))
	for i, sc := range scList {
		nodes[i] = &merkle.Node{
			Position: merkle.NewPosition(0, big.NewInt(0).SetBytes(sc.TreeIndex())),
			Data:     ss.sumStateValue(sc.Value()),
		}
	}
	return nodes
}

func (ss *stateStore) sumStateValue(value []byte) []byte {
	h := sha3.New256()
	h.Write(value)
	return h.Sum(nil)
}

func (ss *stateStore) commitStateChange(sc *core.StateChange) []updateFunc {
	ret := make([]updateFunc, 0)
	ret = append(ret, ss.setState(sc.Key(), sc.Value()))
	ret = append(ret, ss.setTreeIndex(sc.Key(), sc.TreeIndex()))
	return ret
}

func (ss *stateStore) getState(key []byte) ([]byte, error) {
	return getValue(ss.db, concatBytes([]byte{colStateValueByKey}, key))
}

func (ss *stateStore) getMerkleIndex(key []byte) ([]byte, error) {
	return getValue(ss.db, concatBytes([]byte{colMerkleIndexByStateKey}, key))
}

func (ss *stateStore) setState(key, value []byte) updateFunc {
	return func(setter setter) error {
		return setter.Set(
			concatBytes([]byte{colStateValueByKey}, key), value,
		)
	}
}

func (ss *stateStore) setTreeIndex(key, idx []byte) updateFunc {
	return func(setter setter) error {
		return setter.Set(
			concatBytes([]byte{colMerkleIndexByStateKey}, key), idx,
		)
	}
}
