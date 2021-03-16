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

// newQuorumCert creates QC from pb data
func newQuorumCert(data *core_pb.QuorumCert) (*QuorumCert, error) {
	if data == nil {
		return nil, ErrNilQC
	}
	return &QuorumCert{data}, nil
}

// UnmarshalQuorumCert decodes quorum cert from bytes
func UnmarshalQuorumCert(b []byte) (*QuorumCert, error) {
	data := new(core_pb.QuorumCert)
	if err := proto.Unmarshal(b, data); err != nil {
		return nil, err
	}
	return newQuorumCert(data)
}

// Marshal encodes quorum cert as bytes
func (qc *QuorumCert) Marshal() ([]byte, error) {
	return proto.Marshal(qc.data)
}

// Validate godoc
func (qc *QuorumCert) Validate(rs ReplicaStore) error {
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
	if sigs.hasInvalidSig(qc.data.BlockHash) {
		return ErrInvalidSig
	}
	return nil
}
