// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package core

import (
	"errors"

	core_pb "github.com/aungmawjj/juria-blockchain/core/pb"
	"google.golang.org/protobuf/proto"
)

// errors
var (
	ErrNilQC          = errors.New("nil qc")
	ErrNilSig         = errors.New("nil signature")
	ErrNotEnoughSig   = errors.New("not enough signatures in qc")
	ErrDuplicateSig   = errors.New("duplicate signature in qc")
	ErrInvalidSig     = errors.New("invalid signature")
	ErrInvalidReplica = errors.New("voter is not a replica")
)

// QuorumCert type
type QuorumCert struct {
	data *core_pb.QuorumCert
}

// NewQuorumCert creates QC from pb data
func NewQuorumCert(data *core_pb.QuorumCert) *QuorumCert {
	return &QuorumCert{
		data: data,
	}
}

// UnmarshalQuorumCert decodes quorum cert from bytes
func UnmarshalQuorumCert(b []byte) (*QuorumCert, error) {
	data := new(core_pb.QuorumCert)
	if err := proto.Unmarshal(b, data); err != nil {
		return nil, err
	}
	return NewQuorumCert(data), nil
}

// Marshal encodes quorum cert as bytes
func (qc *QuorumCert) Marshal() ([]byte, error) {
	return proto.Marshal(qc.data)
}

// Validate godoc
func (qc *QuorumCert) Validate(rs ReplicaStore) error {
	if qc.data == nil {
		return ErrNilQC
	}
	sigs, err := makeSigList(qc.data.Signatures)
	if err != nil {
		return err
	}
	if len(sigs) < MajorityCount(rs.ReplicaCount()) {
		return ErrNotEnoughSig
	}
	if sigs.hasDuplicate() {
		return ErrDuplicateSig
	}
	if sigs.hasInvalidReplica(rs) {
		return ErrInvalidReplica
	}
	if sigs.hasInvalidSig(qc.data.Block) {
		return ErrInvalidSig
	}
	return nil
}

type sigList []*Signature

func makeSigList(pbsigs []*core_pb.Signature) (sigList, error) {
	sigs := make([]*Signature, len(pbsigs))
	for i, data := range pbsigs {
		if data == nil {
			return nil, ErrNilSig
		}

		var err error
		sigs[i], err = NewSignature(data)
		if err != nil {
			return nil, err
		}
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
