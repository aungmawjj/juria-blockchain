// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package core

import (
	"crypto/ed25519"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlock_Vote(t *testing.T) {
	assert := assert.New(t)

	pub, priv, _ := ed25519.GenerateKey(nil)

	pubKey, err := NewPublicKey(pub)
	assert.NoError(err)

	privKey, err := NewPrivateKey(priv)
	assert.NoError(err)

	block := NewBlock().SetHash([]byte("hash"))

	vote := block.Vote(privKey)
	assert.Equal([]byte("hash"), vote.BlockHash())

	rs := new(MockReplicaStore)
	rs.On("IsReplica", pubKey).Return(true)

	err = vote.Validate(rs)
	assert.NoError(err)
	rs.AssertExpectations(t)

	b, err := vote.Marshal()
	assert.NoError(err)

	vote, err = UnmarshalVote(b)
	assert.NoError(err)

	err = vote.Validate(rs)
	assert.NoError(err)
}
