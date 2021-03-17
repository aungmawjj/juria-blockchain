// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package core

import (
	"crypto/ed25519"
	"testing"

	core_pb "github.com/aungmawjj/juria-blockchain/core/pb"
	"github.com/stretchr/testify/assert"
)

func TestBlock(t *testing.T) {
	assertt := assert.New(t)

	_, priv, _ := ed25519.GenerateKey(nil)
	privKey, _ := NewPrivateKey(priv)

	qc := NewQuorumCert().Build([]*Vote{
		{data: &core_pb.Vote{
			BlockHash: []byte{0},
			Signature: privKey.Sign([]byte{0}).data,
		}},
	})

	blk := NewBlock().
		SetHeight(4).
		SetParentHash([]byte{1}).
		SetProposer([]byte{1}).
		SetExecHeight(0).
		SetQuorumCert(qc).
		SetStateRoot([]byte{1}).
		SetTransactions([][]byte{{1}})

	assertt.Equal(uint64(4), blk.Height())
	assertt.Equal([]byte{1}, blk.ParentHash())
	assertt.Equal([]byte{1}, blk.Proposer())
	assertt.Equal(uint64(0), blk.ExecHeight())
	assertt.Equal(qc, blk.QuorumCert())
	assertt.Equal([]byte{1}, blk.StateRoot())
	assertt.Equal([][]byte{{1}}, blk.Transactions())

	rs := new(MockReplicaStore)
	rs.On("ReplicaCount").Return(1)
	rs.On("IsReplica", privKey.PublicKey()).Return(true)

	assertt.Error(blk.Validate(rs)) // nil hash

	blk.SetHash(blk.Sum())
	bOk, err := blk.Marshal()
	assertt.NoError(err)

	blk.SetQuorumCert(NewQuorumCert())
	blk.SetHash(blk.Sum())
	bNilQC, _ := blk.Marshal()

	blk.SetHash([]byte("invalid hash"))
	bInvalidHash, _ := blk.Marshal()

	// test validate
	tests := []struct {
		name    string
		b       []byte
		wantErr bool
	}{
		{"valid", bOk, false},
		{"nil block", nil, true},
		{"nil qc", bNilQC, true},
		{"invalid", bInvalidHash, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			blk, err := UnmarshalBlock(tt.b)
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

	pub, priv, _ := ed25519.GenerateKey(nil)

	pubKey, err := NewPublicKey(pub)
	assert.NoError(err)

	privKey, err := NewPrivateKey(priv)
	assert.NoError(err)

	blk := NewBlock().SetHash([]byte("hash"))

	vote := blk.Vote(privKey)
	assert.Equal([]byte("hash"), vote.BlockHash())

	rs := new(MockReplicaStore)
	rs.On("IsReplica", pubKey).Return(true)

	err = vote.Validate(rs)
	assert.NoError(err)
	rs.AssertExpectations(t)
}
