// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package core

import (
	"testing"

	"github.com/aungmawjj/juria-blockchain/core/core_pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBlock(t *testing.T) {
	assertt := assert.New(t)

	privKey := GenerateKey(nil)

	qc := NewQuorumCert().Build([]*Vote{
		{data: &core_pb.Vote{
			BlockHash: []byte{0},
			Signature: privKey.Sign([]byte{0}).data,
		}},
	})

	blk := NewBlock().
		SetHeight(4).
		SetParentHash([]byte{1}).
		SetExecHeight(0).
		SetQuorumCert(qc).
		SetMerkleRoot([]byte{1}).
		SetTransactions([][]byte{{1}}).
		Sign(privKey)

	assertt.Equal(uint64(4), blk.Height())
	assertt.Equal([]byte{1}, blk.ParentHash())
	assertt.Equal(privKey.PublicKey(), blk.Proposer())
	assertt.Equal(privKey.PublicKey().Bytes(), blk.data.Proposer)
	assertt.Equal(uint64(0), blk.ExecHeight())
	assertt.Equal(qc, blk.QuorumCert())
	assertt.Equal([]byte{1}, blk.MerkleRoot())
	assertt.Equal([][]byte{{1}}, blk.Transactions())

	rs := new(MockValidatorStore)
	rs.On("ValidatorCount").Return(1)
	rs.On("IsValidator", privKey.PublicKey()).Return(true)
	rs.On("IsValidator", mock.Anything).Return(false)

	bOk, err := blk.Marshal()
	assertt.NoError(err)

	blk.data.Signature = []byte("invalid sig")
	bInvalidSig, _ := blk.Marshal()

	privKey1 := GenerateKey(nil)
	bInvalidValidator, _ := blk.Sign(privKey1).Marshal()

	bNilQC, _ := blk.
		SetQuorumCert(NewQuorumCert()).
		Sign(privKey).
		Marshal()

	blk.data.Hash = []byte("invalid hash")
	bInvalidHash, _ := blk.Marshal()

	// test validate
	tests := []struct {
		name    string
		b       []byte
		wantErr bool
	}{
		{"valid", bOk, false},
		{"invalid sig", bInvalidSig, true},
		{"invalid validator", bInvalidValidator, true},
		{"nil qc", bNilQC, true},
		{"invalid", bInvalidHash, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			blk := NewBlock()
			err := blk.Unmarshal(tt.b)
			assert.NoError(err)

			err = blk.Validate(rs)

			if tt.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestBlock_Vote(t *testing.T) {
	assert := assert.New(t)

	privKey := GenerateKey(nil)

	blk := NewBlock().Sign(privKey)

	vote := blk.Vote(privKey)
	assert.Equal(blk.Hash(), vote.BlockHash())

	rs := new(MockValidatorStore)
	rs.On("IsValidator", privKey.PublicKey()).Return(true)

	err := vote.Validate(rs)
	assert.NoError(err)
	rs.AssertExpectations(t)
}
