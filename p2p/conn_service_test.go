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

func TestConnService_AddPeer(t *testing.T) {
	cs := NewConnService()

	pubKey, _ := core.DecodePublicKey(bytes.Repeat([]byte{1}, ed25519.PublicKeySize))
	pi := NewPeerInfo(pubKey, nil)

	cs.AddPeer(pi)

	assert := assert.New(t)
	p := cs.GetPeer(pi.String())
	if assert.NotNil(p) {
		assert.EqualValues(pi, p.Info())
	}
	assert.Nil(cs.GetPeer("not exist peer"))
}
