// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package core

import (
	"crypto/ed25519"
	"errors"
	"io"

	"github.com/aungmawjj/juria-blockchain/core/core_pb"
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
	return &Signature{
		data: &core_pb.Signature{
			Value:  ed25519.Sign(priv.key, msg),
			PubKey: priv.pubKey.Bytes(),
		},
		pubKey: priv.pubKey,
	}
}

func GenerateKey(rand io.Reader) *PrivateKey {
	_, priv, _ := ed25519.GenerateKey(rand)
	privKey, _ := NewPrivateKey(priv)
	return privKey
}

// Signature type
type Signature struct {
	data   *core_pb.Signature
	pubKey *PublicKey
}

func newSignature(data *core_pb.Signature) (*Signature, error) {
	if data == nil {
		return nil, ErrNilSig
	}
	pubKey, err := NewPublicKey(data.PubKey)
	if err != nil {
		return nil, err
	}
	return &Signature{data, pubKey}, nil
}

// Verify verifies the signature
func (sig *Signature) Verify(msg []byte) bool {
	return ed25519.Verify(sig.pubKey.key, msg, sig.data.Value)
}

// PublicKey returns corresponding public key
func (sig *Signature) PublicKey() *PublicKey {
	return sig.pubKey
}

type sigList []*Signature

func newSigList(pbsigs []*core_pb.Signature) (sigList, error) {
	sigs := make([]*Signature, len(pbsigs))
	for i, data := range pbsigs {
		sig, err := newSignature(data)
		if err != nil {
			return nil, err
		}
		sigs[i] = sig
	}
	return sigs, nil
}

func (sigs sigList) hasDuplicate() bool {
	dmap := make(map[string]struct{}, len(sigs))
	for _, sig := range sigs {
		key := sig.PublicKey().String()
		if _, found := dmap[key]; found {
			return true
		}
		dmap[key] = struct{}{}
	}
	return false
}

func (sigs sigList) hasInvalidValidator(vs ValidatorStore) bool {
	for _, sig := range sigs {
		if !vs.IsValidator(sig.PublicKey()) {
			return true
		}
	}
	return false
}

func (sigs sigList) hasInvalidSig(msg []byte) bool {
	for _, sig := range sigs {
		if !sig.Verify(msg) {
			return true
		}
	}
	return false
}
