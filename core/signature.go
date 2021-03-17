// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package core

import (
	"crypto/ed25519"

	core_pb "github.com/aungmawjj/juria-blockchain/core/pb"
)

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

func (sigs sigList) hasInvalidReplica(rs ReplicaStore) bool {
	for _, sig := range sigs {
		if !rs.IsReplica(sig.PublicKey()) {
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
