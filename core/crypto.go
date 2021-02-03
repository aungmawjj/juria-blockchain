// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package core

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
)

func toBase64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

// PublicKey type
type PublicKey struct {
	key    ed25519.PublicKey
	keyStr string
}

// DecodePublicKey decodes raw bytes to PublicKey
func DecodePublicKey(b []byte) (*PublicKey, error) {
	if len(b) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key size")
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
	b := make([]byte, 0, len(pub.key))
	return append(b, pub.key...)
}

func (pub *PublicKey) String() string {
	return pub.keyStr
}

// Signature type
type Signature struct {
	data   []byte
	pubKey *PublicKey
}

// DecodeSignature decodes raw bytes to signature
func DecodeSignature(b []byte) (*Signature, error) {
	if len(b) != ed25519.PublicKeySize+ed25519.SignatureSize {
		return nil, fmt.Errorf("invalid signature length")
	}
	sig := &Signature{
		data: b[0:ed25519.SignatureSize],
	}
	sig.pubKey, _ = DecodePublicKey(b[ed25519.SignatureSize:])
	return sig, nil
}

// Bytes returns raw bytes
func (sig *Signature) Bytes() []byte {
	return append(sig.data, sig.pubKey.key...)
}

// Verify verifies the signature
func (sig *Signature) Verify(msg []byte) bool {
	return ed25519.Verify(sig.pubKey.key, msg, sig.data)
}

// PublicKey returns corresponding public key
func (sig *Signature) PublicKey() *PublicKey {
	return sig.pubKey
}

// PrivateKey type
type PrivateKey struct {
	key    ed25519.PrivateKey
	pubKey *PublicKey
}

// DecodePrivateKey decodes raw bytes to PrivateKey
func DecodePrivateKey(b []byte) (*PrivateKey, error) {
	if len(b) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key size")
	}
	priv := &PrivateKey{
		key: b,
	}
	priv.pubKey, _ = DecodePublicKey(priv.key.Public().(ed25519.PublicKey))
	return priv, nil
}

// Bytes return raw bytes
func (priv *PrivateKey) Bytes() []byte {
	b := make([]byte, 0, len(priv.key))
	return append(b, priv.key...)
}

// PublicKey returns corresponding public key
func (priv *PrivateKey) PublicKey() *PublicKey {
	return priv.pubKey
}

// Sign signs the message
func (priv *PrivateKey) Sign(msg []byte) *Signature {
	return &Signature{
		ed25519.Sign(priv.key, msg),
		priv.pubKey,
	}
}
