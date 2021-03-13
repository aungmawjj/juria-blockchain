// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package core

import (
	"crypto/ed25519"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignVerify(t *testing.T) {
	assert := assert.New(t)

	pub, priv, err := ed25519.GenerateKey(nil)

	pubKey, err := NewPublicKey(pub)
	assert.NoError(err)

	privKey, err := NewPrivateKey(priv)
	assert.NoError(err)

	msg := []byte("message to be signed")

	sig := privKey.Sign(msg)
	assert.NotNil(sig)

	assert.True(sig.Verify(msg))
	assert.False(sig.Verify([]byte("tampered message")))

	assert.True(pubKey.Equal(sig.PublicKey()))
	assert.True(pubKey.Equal(privKey.PublicKey()))
}
