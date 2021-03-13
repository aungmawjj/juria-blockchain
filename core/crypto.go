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

// Signature type
type Signature struct {
	data   *core_pb.Signature
	pubKey *PublicKey
}

// NewSignature creates signature from pb data
func NewSignature(data *core_pb.Signature) (*Signature, error) {
	sig := &Signature{
		data: data,
	}
	var err error
	sig.pubKey, err = NewPublicKey(data.PubKey)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

// Verify verifies the signature
func (sig *Signature) Verify(msg []byte) bool {
	return ed25519.Verify(sig.pubKey.key, msg, sig.data.Sig)
}

// PublicKey returns corresponding public key
func (sig *Signature) PublicKey() *PublicKey {
	return sig.pubKey
}

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
	sig, _ := NewSignature(
		&core_pb.Signature{
			Sig:    ed25519.Sign(priv.key, msg),
			PubKey: priv.pubKey.Bytes(),
		},
	)
	return sig
}
