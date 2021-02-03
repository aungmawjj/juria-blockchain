// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package core

import (
	"bytes"
	"crypto/ed25519"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodePublicKey(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    []byte
		wantErr bool
	}{
		{
			"valid key",
			bytes.Repeat([]byte{1}, ed25519.PublicKeySize),
			bytes.Repeat([]byte{1}, ed25519.PublicKeySize),
			false,
		},
		{
			"invalid key size",
			bytes.Repeat([]byte{1}, ed25519.PublicKeySize-1),
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodePublicKey(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodePublicKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				assert.EqualValues(t, tt.want, got.Bytes())
				assert.NotEmpty(t, got.String())
				p1, _ := DecodePublicKey(tt.want)
				assert.True(t, got.Equal(p1))
			}
		})
	}
}

func TestSignVerify(t *testing.T) {
	assert := assert.New(t)

	pub, priv, err := ed25519.GenerateKey(nil)

	pubKey, err := DecodePublicKey(pub)
	assert.NoError(err)

	privKey, err := DecodePrivateKey(priv)
	assert.NoError(err)

	msg := []byte("message to be signed")

	sig := privKey.Sign(msg)
	assert.NotNil(sig)

	assert.True(sig.Verify(msg))
	assert.False(sig.Verify([]byte("tampered message")))

	assert.True(pubKey.Equal(sig.PublicKey()))
	assert.True(pubKey.Equal(privKey.PublicKey()))
}
