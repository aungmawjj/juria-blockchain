// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package execution

import (
	"bytes"
	"sync"
)

type StateGetter interface {
	GetState(key []byte) []byte
}

type StateChange struct {
	Key   []byte
	Value []byte
}

// StateTracker tracks state changes in order
// get latest changed state for each key
// get state from base state getter if no changes occured for a key
type StateTracker struct {
	keyPrefix []byte
	getter    StateGetter

	changes     map[string][]byte
	changedKeys [][]byte

	mtx sync.RWMutex
}

func NewStateTracker(getter StateGetter, keyPrefix []byte) *StateTracker {
	return &StateTracker{
		keyPrefix: keyPrefix,
		getter:    getter,

		changes:     make(map[string][]byte),
		changedKeys: make([][]byte, 0),
	}
}

func (trk *StateTracker) GetState(key []byte) []byte {
	trk.mtx.RLock()
	defer trk.mtx.RUnlock()
	return trk.getState(key)
}

func (trk *StateTracker) SetState(key, value []byte) {
	// no lock!
	// SetState must be sequential to maintain consistant order of changes
	trk.setState(key, value)
}

// Spawn creates a new tracker with current tracker as base StateGetter
func (trk *StateTracker) Spawn(keyPrefix []byte) *StateTracker {
	return NewStateTracker(trk, keyPrefix)
}

func (trk *StateTracker) Merge(trk1 *StateTracker) {
	trk.mtx.Lock()
	defer trk.mtx.Unlock()

	trk1.mtx.RLock()
	defer trk1.mtx.RUnlock()

	for _, key := range trk1.changedKeys {
		value := trk1.changes[string(key)]
		trk.setState(key, value)
	}
}

func (trk *StateTracker) GetStateChanges() []*StateChange {
	trk.mtx.RLock()
	defer trk.mtx.RUnlock()

	scList := make([]*StateChange, len(trk.changedKeys))
	for i, key := range trk.changedKeys {
		value := trk.changes[string(key)]
		scList[i] = &StateChange{key, value}
	}
	return scList
}

func (trk *StateTracker) getState(key []byte) []byte {
	key = concatBytes(trk.keyPrefix, key)
	if value, ok := trk.changes[string(key)]; ok {
		return value
	}
	return trk.getter.GetState(key)
}

func (trk *StateTracker) setState(key, value []byte) {
	key = concatBytes(trk.keyPrefix, key)
	keyStr := string(key)
	_, tracked := trk.changes[keyStr]
	trk.changes[keyStr] = value
	if !tracked {
		trk.changedKeys = append(trk.changedKeys, key)
	}
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
