// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package storage

import (
	"bytes"
	"crypto"
	"math/big"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/merkle"
)

type stateStore struct {
	getter   getter
	hashFunc crypto.Hash
}

func (ss *stateStore) loadPrevValues(scList []*core.StateChange) error {
	for _, sc := range scList {
		sc.SetPrevValue(ss.getState(sc.Key()))
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

func (ss *stateStore) setNewTreeIndexes(scList []*core.StateChange, leafCount *big.Int) *big.Int {
	lc := big.NewInt(0).Set(leafCount)
	for _, sc := range scList {
		if sc.PrevTreeIndex() != nil {
			sc.SetTreeIndex(sc.PrevTreeIndex())
		} else {
			idxB := lc.Bytes()
			if len(idxB) == 0 {
				idxB = []byte{0}
			}
			sc.SetTreeIndex(idxB)
			lc.Add(lc, big.NewInt(1))
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
	h := ss.hashFunc.New()
	h.Write(value)
	return h.Sum(nil)
}

func (ss *stateStore) commitStateChanges(scList []*core.StateChange) []updateFunc {
	ret := make([]updateFunc, 0, len(scList))
	for _, sc := range scList {
		ret = append(ret, ss.commitStateChange(sc)...)
	}
	return ret
}

func (ss *stateStore) commitStateChange(sc *core.StateChange) []updateFunc {
	ret := make([]updateFunc, 0)
	ret = append(ret, ss.setState(sc.Key(), sc.Value()))
	if sc.PrevTreeIndex() == nil || !bytes.Equal(sc.PrevTreeIndex(), sc.TreeIndex()) {
		ret = append(ret, ss.setTreeIndex(sc.Key(), sc.TreeIndex()))
	}
	return ret
}

func (ss *stateStore) getState(key []byte) []byte {
	val, err := ss.getter.Get(concatBytes([]byte{colStateValueByKey}, key))
	if err != nil {
		return nil
	}
	return val
}

func (ss *stateStore) getMerkleIndex(key []byte) ([]byte, error) {
	return ss.getter.Get(concatBytes([]byte{colMerkleIndexByStateKey}, key))
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
