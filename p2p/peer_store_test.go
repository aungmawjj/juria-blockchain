// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package p2p

import (
	"bytes"
	"crypto/ed25519"
	"testing"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/stretchr/testify/assert"
)

func TestPeerStore(t *testing.T) {
	assert := assert.New(t)
	s := NewPeerStore()

	// load or store
	pubKey, _ := core.NewPublicKey(bytes.Repeat([]byte{1}, ed25519.PublicKeySize))
	p := NewPeer(pubKey, nil)
	actual, loaded := s.LoadOrStore(p)
	assert.False(loaded)
	assert.Equal(p, actual)

	p1 := NewPeer(pubKey, nil)

	actual, loaded = s.LoadOrStore(p1)
	assert.True(loaded)
	assert.Equal(p, actual)

	// load
	assert.Equal(p, s.Load(pubKey))

	// store
	assert.Equal(p1, s.Store(p1))
	assert.Equal(p1, s.Load(pubKey))

	// list
	assert.Equal([]*Peer{p1}, s.List())

	// delete
	assert.Equal(p1, s.Delete(pubKey))
	assert.Nil(s.Load(pubKey))
	assert.Equal([]*Peer{}, s.List())
}
