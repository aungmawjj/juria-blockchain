// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package storage

import (
	"bytes"
	"crypto"
	"math/big"
	"sort"
	"sync"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/merkle"
)

type stateStore struct {
	getter          getter
	hashFunc        crypto.Hash
	concurrentLimit int
}

func (ss *stateStore) loadPrevValues(scList []*core.StateChange) {
	for _, sc := range scList {
		sc.SetPrevValue(ss.getStateNotFoundNil(sc.Key()))
	}
}

func (ss *stateStore) loadPrevTreeIndexes(scList []*core.StateChange) {
	for _, sc := range scList {
		val, err := ss.getMerkleIndex(sc.Key())
		if err == nil {
			sc.SetPrevTreeIndex(val)
		}
	}
}

func (ss *stateStore) setNewTreeIndexes(scList []*core.StateChange, prevLC *big.Int) *big.Int {
	newKeys := make([]string, 0)
	scByKey := make(map[string]int)
	for i, sc := range scList {
		if sc.PrevTreeIndex() != nil {
			sc.SetTreeIndex(sc.PrevTreeIndex())
		} else {
			key := string(sc.Key())
			newKeys = append(newKeys, key)
			scByKey[key] = i
		}
	}
	sort.Strings(newKeys) // sort new keys to get consistent leaf indexes
	leafCount := big.NewInt(0).Set(prevLC)
	for _, key := range newKeys {
		setLeafIndex(scList[scByKey[key]], leafCount)
		leafCount.Add(leafCount, big.NewInt(1))
	}
	return leafCount
}

func setLeafIndex(sc *core.StateChange, idx *big.Int) {
	idxB := idx.Bytes()
	if len(idxB) == 0 {
		idxB = []byte{0}
	}
	sc.SetTreeIndex(idxB)
}

func (ss *stateStore) computeUpdatedTreeNodes(scList []*core.StateChange) []*merkle.Node {
	nodes := make([]*merkle.Node, len(scList))
	jobs := make(chan int, ss.concurrentLimit)
	defer close(jobs)

	wg := new(sync.WaitGroup)
	for i := 0; i < ss.concurrentLimit; i++ {
		go ss.worker(nodes, scList, jobs, wg)
	}
	for i := range scList {
		wg.Add(1)
		jobs <- i
	}
	wg.Wait()
	return nodes
}

func (ss *stateStore) worker(
	nodes []*merkle.Node, scList []*core.StateChange,
	jobs <-chan int, wg *sync.WaitGroup,
) {
	for i := range jobs {
		sc := scList[i]
		nodes[i] = &merkle.Node{
			Position: merkle.NewPosition(0, big.NewInt(0).SetBytes(sc.TreeIndex())),
			Data:     ss.sumStateValue(sc.Value()),
		}
		wg.Done()
	}
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

func (ss *stateStore) getStateNotFoundNil(key []byte) []byte {
	val, err := ss.getState(key)
	if err != nil {
		return nil
	}
	return val
}

func (ss *stateStore) getState(key []byte) ([]byte, error) {
	return ss.getter.Get(concatBytes([]byte{colStateValueByKey}, key))
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
