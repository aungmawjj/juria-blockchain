// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package consensus

import (
	"errors"
	"testing"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/stretchr/testify/assert"
)

func TestValidator_canVoteProposal(t *testing.T) {
	priv0 := core.GenerateKey(nil)
	priv1 := core.GenerateKey(nil)
	resources := &Resources{
		VldStore: core.NewValidatorStore([]*core.PublicKey{priv0.PublicKey(), priv1.PublicKey()}),
	}

	type given struct {
		execHeight   int
		merkleRoot   []byte
		leaderIdx    int
		verifyTxsErr error
	}
	tests := []struct {
		name     string
		wantErr  bool
		given    given
		proposal *core.Block
	}{
		{"proposer is not leader", true, given{2, []byte{1}, 1, nil},
			core.NewBlock().Sign(priv0)},

		{"different exec height", true, given{2, []byte{1}, 1, nil},
			core.NewBlock().SetExecHeight(3).Sign(priv1)},

		{"different merkle root", true, given{2, []byte{1}, 1, nil},
			core.NewBlock().SetExecHeight(2).SetMerkleRoot([]byte{2}).Sign(priv1)},

		{"verify txs failed", true, given{2, []byte{1}, 1, errors.New("verify txs error")},
			core.NewBlock().SetExecHeight(2).SetMerkleRoot([]byte{1}).Sign(priv1)},

		{"can vote", false, given{2, []byte{1}, 1, nil},
			core.NewBlock().SetExecHeight(2).SetMerkleRoot([]byte{1}).Sign(priv1)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vld := &validator{
				resources: resources,
				state:     newState(resources),
			}
			vld.state.setLeaderIndex(tt.given.leaderIdx)

			storage := new(MockStorage)
			storage.On("GetBlockHeight").Return(tt.given.execHeight)
			storage.On("GetMerkleRoot").Return(tt.given.merkleRoot)
			resources.Storage = storage

			txPool := new(MockTxPool)
			txPool.On("VerifyProposalTxs", tt.proposal.Transactions()).Return(tt.given.verifyTxsErr)
			resources.TxPool = txPool

			assert := assert.New(t)
			if tt.wantErr {
				assert.Error(vld.canVoteProposal(tt.proposal))
			} else {
				assert.NoError(vld.canVoteProposal(tt.proposal))
			}
		})
	}
}
