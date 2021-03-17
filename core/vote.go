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

func NewVote() *Vote {
	return &Vote{
		data: new(core_pb.Vote),
	}
}

// Validate vote
func (vote *Vote) Validate(rs ReplicaStore) error {
	if vote.data == nil {
		return ErrNilVote
	}
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

func (vote *Vote) setData(data *core_pb.Vote) *Vote {
	vote.data = data
	return vote
}

// BlockHash of vote
func (vote *Vote) BlockHash() []byte { return vote.data.BlockHash }

// Marshal encodes vote as bytes
func (vote *Vote) Marshal() ([]byte, error) {
	return proto.Marshal(vote.data)
}

// UnmarshalVote decodes vote from bytes
func UnmarshalVote(b []byte) (*Vote, error) {
	data := new(core_pb.Vote)
	if err := proto.Unmarshal(b, data); err != nil {
		return nil, err
	}
	return NewVote().setData(data), nil
}
