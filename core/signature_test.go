// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignVerify(t *testing.T) {
	assert := assert.New(t)

	privKey := GenerateKey(nil)

	assert.Equal(privKey.PublicKey(), privKey.PublicKey())

	msg := []byte("message to be signed")

	sig := privKey.Sign(msg)
	assert.NotNil(sig)

	assert.True(sig.Verify(msg))
	assert.False(sig.Verify([]byte("tampered message")))

	assert.Equal(privKey.PublicKey(), sig.PublicKey())
}
