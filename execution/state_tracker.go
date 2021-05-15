// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package execution

import (
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
	getter StateGetter

	changes     map[string][]byte
	changedKeys [][]byte

	mtx sync.RWMutex
}

func NewStateTracker(getter StateGetter) *StateTracker {
	return &StateTracker{
		getter:      getter,
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
	trk.mtx.Lock()
	defer trk.mtx.Unlock()
	trk.setState(key, value)
}

// Spawn creates a new tracker with current tracker as base StateGetter
func (trk *StateTracker) Spawn() *StateTracker {
	return NewStateTracker(trk)
}

func (trk *StateTracker) Merge(trk1 *StateTracker) {
	trk.mtx.Lock()
	defer trk.mtx.Unlock()

	trk1.mtx.RLock()
	defer trk1.mtx.RUnlock()

	for _, key := range trk1.changedKeys {
		trk.setState(key, trk1.getState(key))
	}
}

func (trk *StateTracker) GetStateChanges() []*StateChange {
	trk.mtx.RLock()
	defer trk.mtx.RUnlock()

	scList := make([]*StateChange, len(trk.changedKeys))
	for i, key := range trk.changedKeys {
		value := trk.getState(key)
		scList[i] = &StateChange{key, value}
	}
	return scList
}

func (trk *StateTracker) getState(key []byte) []byte {
	if value, ok := trk.changes[string(key)]; ok {
		return value
	}
	return trk.getter.GetState(key)
}

func (trk *StateTracker) setState(key, value []byte) {
	keyStr := string(key)
	_, tracked := trk.changes[keyStr]
	trk.changes[keyStr] = value
	if !tracked {
		trk.changedKeys = append(trk.changedKeys, key)
	}
}
