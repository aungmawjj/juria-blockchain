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

func TestPeerStore_LoadOrStore(t *testing.T) {
	assert := assert.New(t)
	s := newPeerStore()

	pubKey, _ := core.DecodePublicKey(bytes.Repeat([]byte{1}, ed25519.PublicKeySize))
	p := NewPeer(pubKey, nil)
	actual, loaded := s.LoadOrStore(p)
	assert.False(loaded)
	assert.True(p == actual)

	p1 := NewPeer(pubKey, nil)

	actual, loaded = s.LoadOrStore(p1)
	assert.True(loaded)
	assert.False(p1 == actual)
}
