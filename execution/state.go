// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package execution

import (
	"bytes"
	"sort"
	"sync"
	"time"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/logger"
)

type StateRO interface {
	VerifyState(key []byte) ([]byte, error)
	GetState(key []byte) []byte
}

type State interface {
	StateRO
	SetState(key, value []byte)
}

// stateTracker tracks state changes in key order
// get latest changed state for each key
// get state from base state getter if no changes occured for a key
type stateTracker struct {
	keyPrefix []byte
	baseState StateRO

	trackDep     bool
	dependencies map[string]struct{} // getState/verifyState calls
	changes      map[string][]byte   // setState calls

	mtxChg sync.RWMutex
	mtxDep sync.RWMutex
}

var _ State = (*stateTracker)(nil)

func newStateTracker(state StateRO, keyPrefix []byte) *stateTracker {
	return &stateTracker{
		keyPrefix: keyPrefix,
		baseState: state,

		dependencies: make(map[string]struct{}),
		changes:      make(map[string][]byte),
	}
}

func (trk *stateTracker) VerifyState(key []byte) ([]byte, error) {
	trk.mtxChg.RLock()
	defer trk.mtxChg.RUnlock()
	return trk.verifyState(key)
}

func (trk *stateTracker) GetState(key []byte) []byte {
	trk.mtxChg.RLock()
	defer trk.mtxChg.RUnlock()
	return trk.getState(key)
}

func (trk *stateTracker) SetState(key, value []byte) {
	trk.mtxChg.Lock()
	defer trk.mtxChg.Unlock()
	trk.setState(key, value)
}

// spawn creates a new tracker with current tracker as base StateGetter
func (trk *stateTracker) spawn(keyPrefix []byte) *stateTracker {
	child := newStateTracker(trk, keyPrefix)
	child.trackDep = true
	return child
}

func (trk *stateTracker) hasDependencyChanges(child *stateTracker) bool {
	trk.mtxChg.RLock()
	defer trk.mtxChg.RUnlock()

	child.mtxDep.RLock()
	defer child.mtxDep.RUnlock()

	prefixStr := string(trk.keyPrefix)

	for key := range child.dependencies {
		key = prefixStr + key
		if _, changed := trk.changes[key]; changed {
			return true
		}
	}
	return false
}

func (trk *stateTracker) merge(child *stateTracker) {
	trk.mtxChg.Lock()
	defer trk.mtxChg.Unlock()

	child.mtxChg.RLock()
	defer child.mtxChg.RUnlock()

	for key, value := range child.changes {
		trk.setState([]byte(key), value)
	}
}

func (trk *stateTracker) getStateChanges() []*core.StateChange {
	trk.mtxChg.RLock()
	defer trk.mtxChg.RUnlock()

	start := time.Now()
	keys := make([]string, 0, len(trk.changes))
	for key := range trk.changes {
		keys = append(keys, key)
	}
	// state changes are sorted by keys to keep it consistant across different nodes
	// if required, performance improvement should be done
	// 1. sort only new merkle leaf-nodes in storage
	sort.Strings(keys)
	elapsed := time.Since(start)
	if elapsed > 1*time.Millisecond {
		logger.I().Debugw("sorted state changes", "count", len(keys), "elapsed", elapsed)
	}
	scList := make([]*core.StateChange, len(keys))
	for i, key := range keys {
		value := trk.changes[key]
		scList[i] = core.NewStateChange().SetKey([]byte(key)).SetValue(value)
	}
	return scList
}

func (trk *stateTracker) verifyState(key []byte) ([]byte, error) {
	key = concatBytes(trk.keyPrefix, key)
	trk.setDependency(key)
	if value, ok := trk.changes[string(key)]; ok {
		return value, nil
	}
	return trk.baseState.VerifyState(key)
}

func (trk *stateTracker) getState(key []byte) []byte {
	key = concatBytes(trk.keyPrefix, key)
	trk.setDependency(key)
	if value, ok := trk.changes[string(key)]; ok {
		return value
	}
	return trk.baseState.GetState(key)
}

func (trk *stateTracker) setDependency(key []byte) {
	if !trk.trackDep {
		return
	}
	trk.mtxDep.Lock()
	defer trk.mtxDep.Unlock()
	trk.dependencies[string(key)] = struct{}{}
}

func (trk *stateTracker) setState(key, value []byte) {
	key = concatBytes(trk.keyPrefix, key)
	keyStr := string(key)
	trk.changes[keyStr] = value
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
