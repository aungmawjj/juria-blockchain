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

func TestQuorumCert(t *testing.T) {

	privKeys := make([]*PrivateKey, 5)

	rs := new(MockReplicaStore)
	rs.On("ReplicaCount").Return(4)

	for i := range privKeys {
		_, priv, _ := ed25519.GenerateKey(nil)
		privKeys[i], _ = NewPrivateKey(priv)
		if i != 4 {
			rs.On("IsReplica", privKeys[i].pubKey).Return(true)
		}
	}
	rs.On("IsReplica", mock.Anything).Return(false)

	blockHash := []byte{1}
	votes := make([]*Vote, len(privKeys))
	for i, priv := range privKeys {
		votes[i] = &Vote{&core_pb.Vote{
			BlockHash: blockHash,
			Signature: priv.Sign(blockHash).data,
		}}
	}

	nilSigVote := &Vote{&core_pb.Vote{
		BlockHash: blockHash,
		Signature: nil,
	}}

	invalidSigVote := &Vote{&core_pb.Vote{
		BlockHash: blockHash,
		Signature: privKeys[0].Sign(blockHash).data,
	}}

	qc := NewQuorumCert().Build([]*Vote{votes[2], votes[1], votes[0]})
	qcValid, err := qc.Marshal()
	assert.NoError(t, err)

	qc = NewQuorumCert().Build([]*Vote{votes[2], votes[1], votes[0], votes[3]})
	qcValidFull, _ := qc.Marshal()

	qc = NewQuorumCert().Build([]*Vote{votes[1], votes[0]})
	qcNotEnoughSig, _ := qc.Marshal()

	qc = NewQuorumCert().Build([]*Vote{votes[2], votes[3], nilSigVote, votes[0]})
	qcNilSig, _ := qc.Marshal()

	qc = NewQuorumCert().Build([]*Vote{votes[2], votes[3], votes[0], votes[2]})
	qcDuplicateKey, _ := qc.Marshal()

	qc = NewQuorumCert().Build([]*Vote{votes[1], votes[3], votes[0], votes[4]})
	qcInvalidReplica, _ := qc.Marshal()

	qc = NewQuorumCert().Build([]*Vote{votes[2], votes[1], votes[4], invalidSigVote})
	qcInvalidSig, _ := qc.Marshal()

	// test validate
	tests := []struct {
		name    string
		b       []byte
		wantErr bool
	}{
		{"valid", qcValid, false},
		{"valid full", qcValidFull, false},
		{"nil qc", nil, true},
		{"not enough sig", qcNotEnoughSig, true},
		{"nil sig", qcNilSig, true},
		{"duplicate key", qcDuplicateKey, true},
		{"invalid replica", qcInvalidReplica, true},
		{"invalid sig", qcInvalidSig, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			qc, err := UnmarshalQuorumCert(tt.b)
			assert.NoError(err)

			err = qc.Validate(rs)

			if tt.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}
