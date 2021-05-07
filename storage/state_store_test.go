// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package storage

import (
	"testing"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/stretchr/testify/assert"
)

func TestStateStore_updateState(t *testing.T) {
	assert := assert.New(t)

	db := createOnMemoryDB()
	ss := &StateStore{db}

	upd := core.NewStateChange().
		SetKey([]byte{1}).
		SetValue([]byte{2}).
		SetTreeIndex([]byte{1})

	_, err := ss.GetState(upd.Key())
	assert.Error(err)

	fn := ss.updateState(upd)
	updateDB(db, []updateFunc{fn})

	val, err := ss.GetState(upd.Key())
	assert.NoError(err)
	assert.Equal(upd.Value(), val)

	idx, err := ss.getMerkleIndex(upd.Key())
	assert.NoError(err)
	assert.Equal(upd.TreeIndex(), idx)
}
