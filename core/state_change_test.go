// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStateChange(t *testing.T) {
	assert := assert.New(t)

	sc := NewStateChange().
		SetKey([]byte("key")).
		SetValue([]byte("value")).
		SetPrevValue([]byte("prevValue")).
		SetTreeIndex([]byte{1}).
		SetPrevTreeIndex(nil)

	b, err := sc.Marshal()
	assert.NoError(err)

	sc = NewStateChange()
	err = sc.Unmarshal(b)
	assert.NoError(err)

	assert.Equal([]byte("key"), sc.Key())
	assert.Equal([]byte("value"), sc.Value())
	assert.Equal([]byte("prevValue"), sc.PrevValue())
	assert.Equal([]byte{1}, sc.TreeIndex())
	assert.Nil(sc.PrevTreeIndex())
}
