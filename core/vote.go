// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package core

import (
	core_pb "github.com/aungmawjj/juria-blockchain/core/pb"
	"google.golang.org/protobuf/proto"
)

// Vote type
type Vote struct {
	data *core_pb.Vote
}

// NewVote creates vote from pb data
func NewVote(data *core_pb.Vote) *Vote {
	return &Vote{
		data: data,
	}
}

// UnmarshalVote decodes vote from bytes
func UnmarshalVote(b []byte) (*Vote, error) {
	data := new(core_pb.Vote)
	if err := proto.Unmarshal(b, data); err != nil {
		return nil, err
	}
	return NewVote(data), nil
}

// Marshal encodes vote as bytes
func (vote *Vote) Marshal() ([]byte, error) {
	return proto.Marshal(vote.data)
}

// Validate vote
func (vote *Vote) Validate(rs ReplicaStore) error {
	if vote.data.Signature == nil {
		return ErrNilSig
	}
	sig, err := NewSignature(vote.data.Signature)
	if err != nil {
		return err
	}
	if !rs.IsReplica(sig.PublicKey()) {
		return ErrInvalidReplica
	}
	if !sig.Verify(vote.data.Block) {
		return ErrInvalidSig
	}
	return nil
}
