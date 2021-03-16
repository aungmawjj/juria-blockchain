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
	ErrNilVote = errors.New("nil vote")
)

// Vote type
type Vote struct {
	data *core_pb.Vote
}

// newVote creates vote from pb data
func newVote(data *core_pb.Vote) (*Vote, error) {
	if data == nil {
		return nil, ErrNilVote
	}
	return &Vote{data}, nil
}

// UnmarshalVote decodes vote from bytes
func UnmarshalVote(b []byte) (*Vote, error) {
	data := new(core_pb.Vote)
	if err := proto.Unmarshal(b, data); err != nil {
		return nil, err
	}
	return newVote(data)
}

// Marshal encodes vote as bytes
func (vote *Vote) Marshal() ([]byte, error) {
	return proto.Marshal(vote.data)
}

// Validate vote
func (vote *Vote) Validate(rs ReplicaStore) error {
	sig, err := newSignature(vote.data.Signature)
	if err != nil {
		return err
	}
	if !rs.IsReplica(sig.PublicKey()) {
		return ErrInvalidReplica
	}
	if !sig.Verify(vote.data.BlockHash) {
		return ErrInvalidSig
	}
	return nil
}

// BlockHash of vote
func (vote *Vote) BlockHash() []byte { return vote.data.BlockHash }
