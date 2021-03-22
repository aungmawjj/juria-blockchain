// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package core

import (
	"crypto/ed25519"
	"testing"

	core_pb "github.com/aungmawjj/juria-blockchain/core/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestVote(t *testing.T) {
	vote := &Vote{
		data: &core_pb.Vote{},
	}
	vNilSig, err := vote.Marshal()
	assert.NoError(t, err)

	_, priv, _ := ed25519.GenerateKey(nil)
	proposer, _ := NewPrivateKey(priv)

	_, priv, _ = ed25519.GenerateKey(nil)
	privKey, _ := NewPrivateKey(priv)

	blk := NewBlock().Sign(proposer)
	vote = blk.Vote(privKey)
	vOk, _ := vote.Marshal()

	vote.data.BlockHash = []byte("invalid hash")
	vInvalid, _ := vote.Marshal()

	// test validate
	tests := []struct {
		name    string
		b       []byte
		wantErr bool
	}{
		{"valid", vOk, false},
		{"nil vote", nil, true},
		{"nil signature", vNilSig, true},
		{"invalid", vInvalid, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			vote, err := UnmarshalVote(tt.b)
			assert.NoError(err)

			rs := new(MockValidatorStore)
			rs.On("IsValidator", mock.Anything).Return(true)

			err = vote.Validate(rs)

			if tt.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}
