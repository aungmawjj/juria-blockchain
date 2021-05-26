// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package consensus

import (
	"testing"
	"time"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/hotstuff"
	"github.com/stretchr/testify/assert"
)

func setupPacemaker() (*pacemaker, *core.Block) {
	key1 := core.GenerateKey(nil)
	key2 := core.GenerateKey(nil)
	vlds := []*core.PublicKey{
		key1.PublicKey(),
		key2.PublicKey(),
	}
	resources := &Resources{
		VldStore: core.NewValidatorStore(vlds),
	}

	b0 := core.NewBlock().Sign(key1)
	q0 := core.NewQuorumCert().Build([]*core.Vote{b0.ProposerVote()})
	b0.SetQuorumCert(q0)

	state := newState(resources)
	state.setBlock(b0)
	hsDriver := &hsDriver{
		resources: resources,
		state:     state,
	}
	hotstuff := hotstuff.New(hsDriver, newHsBlock(b0, state), newHsQC(q0, state))
	return &pacemaker{
		resources: resources,
		config:    DefaultConfig,
		state:     state,
		hotstuff:  hotstuff,
	}, b0
}

func TestPacemaker_changeView(t *testing.T) {
	assert := assert.New(t)

	pm, b0 := setupPacemaker()
	pm.state.setLeaderIndex(1)

	msgSvc := new(MockMsgService)
	msgSvc.On("SendNewView", pm.resources.VldStore.GetValidator(0), b0.QuorumCert()).Return(nil)
	pm.resources.MsgSvc = msgSvc

	pm.changeView()

	msgSvc.AssertExpectations(t)
	assert.True(pm.getPendingViewChange())
	assert.EqualValues(pm.state.getLeaderIndex(), 0)
}

func Test_pacemaker_needViewTimerResetForNewQC(t *testing.T) {
	assert := assert.New(t)

	pm1, _ := setupPacemaker()
	pm2, _ := setupPacemaker()

	pm1.setPendingViewChange(true)
	pm2.setPendingViewChange(false)

	tests := []struct {
		name        string
		pm          *pacemaker
		proposerIdx int
		want        bool
	}{
		{"pending and same leader", pm1, 0, true},
		{"not pending and different leader", pm2, 1, true},
		{"pending and different leader", pm1, 1, false},
		{"not pending and same leader", pm2, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.EqualValues(tt.want, tt.pm.needViewTimerResetForNewQC(tt.proposerIdx))
		})
	}
}

func TestPacemaker_resetViewTimer(t *testing.T) {
	assert := assert.New(t)

	pm, _ := setupPacemaker()
	pm.viewTimer = time.NewTimer(pm.config.ViewWidth)
	pm.setPendingViewChange(true)

	pm.resetViewTimer(1)

	assert.False(pm.getPendingViewChange())
	assert.EqualValues(pm.state.getLeaderIndex(), 1)
}
