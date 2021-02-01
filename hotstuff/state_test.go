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
	assert.Equal(b0, s.getBExec())
	assert.Equal(b0, s.getBLock())
	assert.Equal(b0, s.getBLeaf())
	assert.Equal(q0, s.getQCHigh())
	assert.Equal(b0.Height(), s.getVHeight())
}
