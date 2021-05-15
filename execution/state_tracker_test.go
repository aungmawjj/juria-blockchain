// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package execution

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type mapStateStore struct {
	stateMap map[string][]byte
}

func newMapStateStore() *mapStateStore {
	return &mapStateStore{
		stateMap: make(map[string][]byte),
	}
}

func (store *mapStateStore) GetState(key []byte) []byte {
	return store.stateMap[string(key)]
}

func (store *mapStateStore) setState(key, value []byte) {
	store.stateMap[string(key)] = value
}

func TestStateTracker_GetState(t *testing.T) {
	assert := assert.New(t)

	ms := newMapStateStore()
	trk := NewStateTracker(ms, nil)
	ms.setState([]byte{1}, []byte{200})

	assert.Equal([]byte{200}, trk.GetState([]byte{1}))
	assert.Nil(trk.GetState([]byte{2}))

	trkChild := trk.Spawn(nil)
	assert.Equal([]byte{200}, trkChild.GetState([]byte{1}), "child get state from root store")
	assert.Nil(trkChild.GetState([]byte{2}))

	trk.SetState([]byte{1}, []byte{100})
	assert.Equal([]byte{100}, trk.GetState([]byte{1}), "get latest state")
	assert.Equal([]byte{100}, trkChild.GetState([]byte{1}), "child get latest state from parent")
}

func TestStateTracker_SetState(t *testing.T) {
	assert := assert.New(t)

	ms := newMapStateStore()
	trk := NewStateTracker(ms, nil)
	ms.setState([]byte{1}, []byte{200})

	trk.SetState([]byte{1}, []byte{100})
	trk.SetState([]byte{1}, []byte{50})

	assert.Equal([]byte{50}, trk.GetState([]byte{1}))
	scList := trk.GetStateChanges()
	assert.Equal(1, len(scList))
	assert.Equal([]byte{1}, scList[0].Key)
	assert.Equal([]byte{50}, scList[0].Value)

	trk.SetState([]byte{3}, []byte{30})
	trk.SetState([]byte{2}, []byte{60})
	trk.setState([]byte{2}, []byte{20})
	trk.SetState([]byte{1}, []byte{10})

	assert.Equal([]byte{10}, trk.GetState([]byte{1}))
	assert.Equal([]byte{30}, trk.GetState([]byte{3}))
	assert.Equal([]byte{20}, trk.GetState([]byte{2}))

	scList = trk.GetStateChanges()
	assert.Equal(3, len(scList))
	assert.Equal([]byte{1}, scList[0].Key)
	assert.Equal([]byte{3}, scList[1].Key)
	assert.Equal([]byte{2}, scList[2].Key)

	assert.Equal([]byte{10}, scList[0].Value)
	assert.Equal([]byte{30}, scList[1].Value)
	assert.Equal([]byte{20}, scList[2].Value)
}

func TestStateTracker_Merge(t *testing.T) {
	assert := assert.New(t)

	ms := newMapStateStore()
	trk := NewStateTracker(ms, nil)

	trk.SetState([]byte{1}, []byte{200})
	trkChild := trk.Spawn(nil)
	trkChild.SetState([]byte{2}, []byte{20})
	trkChild.SetState([]byte{1}, []byte{10})

	assert.Equal([]byte{20}, trkChild.GetState([]byte{2}))
	assert.Equal([]byte{10}, trkChild.GetState([]byte{1}))

	assert.Equal([]byte{200}, trk.GetState([]byte{1}), "child does not set parent state")

	trk.Merge(trkChild)

	assert.Equal([]byte{10}, trk.GetState([]byte{1}))
	assert.Equal([]byte{20}, trk.GetState([]byte{2}))

	scList := trk.GetStateChanges()
	assert.Equal(2, len(scList))
	assert.Equal([]byte{1}, scList[0].Key, "change order should be in favor of parent tracker")
	assert.Equal([]byte{2}, scList[1].Key)
}

func TestStateTracker_WithPrefix(t *testing.T) {
	assert := assert.New(t)

	ms := newMapStateStore()
	trk := NewStateTracker(ms, nil)
	trk.SetState([]byte{1, 1}, []byte{50})

	trkChild := trk.Spawn([]byte{1})
	assert.Equal([]byte{50}, trkChild.GetState([]byte{1}))

	trkChild.SetState([]byte{1}, []byte{10})
	trkChild.SetState([]byte{2}, []byte{20})
	assert.Equal([]byte{10}, trkChild.GetState([]byte{1}))
	assert.Equal([]byte{20}, trkChild.GetState([]byte{2}))

	scList := trkChild.GetStateChanges()
	assert.Equal(2, len(scList))
	assert.Equal([]byte{1, 1}, scList[0].Key)
	assert.Equal([]byte{1, 2}, scList[1].Key)

	trk.Merge(trkChild)
	assert.Equal([]byte{10}, trk.GetState([]byte{1, 1}))
	assert.Equal([]byte{20}, trk.GetState([]byte{1, 2}))
}
