// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package execution

import (
	"bytes"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/aungmawjj/juria-blockchain/core"
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

	changes map[string][]byte

	mtx sync.RWMutex
}

var _ State = (*stateTracker)(nil)

func newStateTracker(state StateRO, keyPrefix []byte) *stateTracker {
	return &stateTracker{
		keyPrefix: keyPrefix,
		baseState: state,

		changes: make(map[string][]byte),
	}
}

func (trk *stateTracker) VerifyState(key []byte) ([]byte, error) {
	trk.mtx.RLock()
	defer trk.mtx.RUnlock()
	return trk.verifyState(key)
}

func (trk *stateTracker) GetState(key []byte) []byte {
	trk.mtx.RLock()
	defer trk.mtx.RUnlock()
	return trk.getState(key)
}

func (trk *stateTracker) SetState(key, value []byte) {
	trk.mtx.Lock()
	defer trk.mtx.Unlock()
	trk.setState(key, value)
}

// spawn creates a new tracker with current tracker as base StateGetter
func (trk *stateTracker) spawn(keyPrefix []byte) *stateTracker {
	return newStateTracker(trk, keyPrefix)
}

func (trk *stateTracker) merge(trk1 *stateTracker) {
	trk.mtx.Lock()
	defer trk.mtx.Unlock()

	trk1.mtx.RLock()
	defer trk1.mtx.RUnlock()

	for key, value := range trk1.changes {
		trk.setState([]byte(key), value)
	}
}

func (trk *stateTracker) getStateChanges() []*core.StateChange {
	trk.mtx.RLock()
	defer trk.mtx.RUnlock()

	start := time.Now()
	keys := make([]string, 0, len(trk.changes))
	for key := range trk.changes {
		keys = append(keys, key)
	}
	// scList is sorted by keys to keep it consistant across different nodes
	sort.Strings(keys)
	scList := make([]*core.StateChange, len(keys))
	for i, key := range keys {
		value := trk.changes[key]
		scList[i] = core.NewStateChange().SetKey([]byte(key)).SetValue(value)
	}
	if len(scList) > 0 {
		fmt.Println("get sorted sc list", time.Since(start))
	}
	return scList
}

func (trk *stateTracker) verifyState(key []byte) ([]byte, error) {
	key = concatBytes(trk.keyPrefix, key)
	if value, ok := trk.changes[string(key)]; ok {
		return value, nil
	}
	return trk.baseState.VerifyState(key)
}

func (trk *stateTracker) getState(key []byte) []byte {
	key = concatBytes(trk.keyPrefix, key)
	if value, ok := trk.changes[string(key)]; ok {
		return value
	}
	return trk.baseState.GetState(key)
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
