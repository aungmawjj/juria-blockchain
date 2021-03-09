// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package hotstuff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_state_init(t *testing.T) {

	b0 := new(MockBlock)
	q0 := new(MockQC)

	s := &state{}

	b0.On("Height").Return(10)
	s.init(b0, q0)

	assert := assert.New(t)
	assert.Equal(b0, s.GetBExec())
	assert.Equal(b0, s.GetBLock())
	assert.Equal(b0, s.GetBLeaf())
	assert.Equal(q0, s.GetQCHigh())
	assert.Equal(b0.Height(), s.GetVHeight())
}

func Test_state_GetVotes(t *testing.T) {
	s := &state{}

	assert := assert.New(t)
	assert.Equal([]Vote{}, s.GetVotes())

	v0 := new(MockVote)
	s.votes = map[string]Vote{"r0": v0}
	assert.Equal([]Vote{v0}, s.GetVotes())
}
