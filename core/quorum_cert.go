// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package core

import (
	"errors"

	"github.com/aungmawjj/juria-blockchain/core/core_pb"
	"google.golang.org/protobuf/proto"
)

// errors
var (
	ErrNilQC            = errors.New("nil qc")
	ErrNilSig           = errors.New("nil signature")
	ErrNotEnoughSig     = errors.New("not enough signatures in qc")
	ErrDuplicateSig     = errors.New("duplicate signature in qc")
	ErrInvalidSig       = errors.New("invalid signature")
	ErrInvalidValidator = errors.New("voter is not a validator")
)

// QuorumCert type
type QuorumCert struct {
	data *core_pb.QuorumCert
}

func NewQuorumCert() *QuorumCert {
	return &QuorumCert{
		data: new(core_pb.QuorumCert),
	}
}

// Validate godoc
func (qc *QuorumCert) Validate(vs ValidatorStore) error {
	if qc.data == nil {
		return ErrNilQC
	}
	sigs, err := newSigList(qc.data.Signatures)
	if err != nil {
		return err
	}
	if len(sigs) < MajorityCount(vs.ValidatorCount()) {
		return ErrNotEnoughSig
	}
	if sigs.hasDuplicate() {
		return ErrDuplicateSig
	}
	if sigs.hasInvalidValidator(vs) {
		return ErrInvalidValidator
	}
	if sigs.hasInvalidSig(qc.data.BlockHash) {
		return ErrInvalidSig
	}
	return nil
}

// newQuorumCert creates QC from pb data
func (qc *QuorumCert) setData(data *core_pb.QuorumCert) *QuorumCert {
	qc.data = data
	return qc
}

func (qc *QuorumCert) Build(votes []*Vote) *QuorumCert {
	qc.data.BlockHash = votes[0].data.BlockHash
	qc.data.Signatures = make([]*core_pb.Signature, len(votes))
	for i, vote := range votes {
		qc.data.Signatures[i] = vote.data.Signature
	}
	return qc
}

func (qc *QuorumCert) BlockHash() []byte { return qc.data.BlockHash }

// Marshal encodes quorum cert as bytes
func (qc *QuorumCert) Marshal() ([]byte, error) {
	return proto.Marshal(qc.data)
}

// UnmarshalQuorumCert decodes quorum cert from bytes
func UnmarshalQuorumCert(b []byte) (*QuorumCert, error) {
	data := new(core_pb.QuorumCert)
	if err := proto.Unmarshal(b, data); err != nil {
		return nil, err
	}
	return NewQuorumCert().setData(data), nil
}
