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
	pi := &PeerInfo{pubKey: pubKey}

	var p Peer
	p = newPeerConnecting(pi)
	_, loaded := s.LoadOrStore(p.String(), p)
	assert.False(loaded)

	p = newPeerConnected(pi, newRWCPipe(), func(p Peer) {})
	actual, loaded := s.LoadOrStore(p.String(), p)
	assert.True(loaded)
	assert.NotEqual(p, actual)
}
