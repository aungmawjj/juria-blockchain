// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package core

import (
	"crypto/ed25519"
	"errors"

	core_pb "github.com/aungmawjj/juria-blockchain/core/pb"
)

// errors
var (
	ErrInvalidKeySize = errors.New("invalid key size")
)

// PublicKey type
type PublicKey struct {
	key    ed25519.PublicKey
	keyStr string
}

// NewPublicKey creates PublicKey from bytes
func NewPublicKey(b []byte) (*PublicKey, error) {
	if len(b) != ed25519.PublicKeySize {
		return nil, ErrInvalidKeySize
	}
	return &PublicKey{
		key:    b,
		keyStr: toBase64(b),
	}, nil
}

// Equal checks whether pub and x has the same value
func (pub *PublicKey) Equal(x *PublicKey) bool {
	return pub.key.Equal(x.key)
}

// Bytes return raw bytes
func (pub *PublicKey) Bytes() []byte {
	return pub.key
}

func (pub *PublicKey) String() string {
	return pub.keyStr
}

// PrivateKey type
type PrivateKey struct {
	key    ed25519.PrivateKey
	pubKey *PublicKey
}

// NewPrivateKey creates PrivateKey from bytes
func NewPrivateKey(b []byte) (*PrivateKey, error) {
	if len(b) != ed25519.PrivateKeySize {
		return nil, ErrInvalidKeySize
	}
	priv := &PrivateKey{
		key: b,
	}
	priv.pubKey, _ = NewPublicKey(priv.key.Public().(ed25519.PublicKey))
	return priv, nil
}

// Bytes return raw bytes
func (priv *PrivateKey) Bytes() []byte {
	return priv.key
}

// PublicKey returns corresponding public key
func (priv *PrivateKey) PublicKey() *PublicKey {
	return priv.pubKey
}

// Sign signs the message
func (priv *PrivateKey) Sign(msg []byte) *Signature {
	sig, _ := newSignature(
		&core_pb.Signature{
			Value:  ed25519.Sign(priv.key, msg),
			PubKey: priv.pubKey.Bytes(),
		},
	)
	return sig
}
