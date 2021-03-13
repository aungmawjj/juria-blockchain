// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package core

import (
	"bytes"
	"errors"

	core_pb "github.com/aungmawjj/juria-blockchain/core/pb"
	"golang.org/x/crypto/sha3"
	"google.golang.org/protobuf/proto"
)

// errors
var (
	ErrInvalidBlockHash = errors.New("invalid block hash")
)

// Block type
type Block struct {
	data       *core_pb.Block
	quorumCert *QuorumCert
}

// NewBlock creates Block from pb data
func NewBlock(data *core_pb.Block) *Block {
	blk := &Block{
		data: data,
	}
	blk.quorumCert = NewQuorumCert(data.QuorumCert)
	return blk
}

// UnmarshalBlock decodes block from bytes
func UnmarshalBlock(b []byte) (*Block, error) {
	data := new(core_pb.Block)
	if err := proto.Unmarshal(b, data); err != nil {
		return nil, err
	}
	return NewBlock(data), nil
}

// Marshal encodes blk as bytes
func (blk *Block) Marshal() ([]byte, error) {
	return proto.Marshal(blk.data)
}

// Sum returns sha3 sum of block
func (blk *Block) Sum() []byte {
	h := sha3.New256()
	h.Write(uint64ToBytes(blk.data.Height))
	h.Write(blk.data.Parent)
	h.Write(blk.data.Proposer)
	if blk.data.QuorumCert != nil {
		h.Write(blk.data.QuorumCert.Block) // qc reference block hash
	}
	h.Write(uint64ToBytes(blk.data.ExecHeight))
	h.Write(blk.data.StateRoot)
	for _, txHash := range blk.data.Transactions {
		h.Write(txHash)
	}
	return h.Sum(nil)
}

// Validate block
func (blk *Block) Validate(rs ReplicaStore) error {
	if err := blk.quorumCert.Validate(rs); err != nil {
		return err
	}
	if !bytes.Equal(blk.Sum(), blk.data.Hash) {
		return ErrInvalidBlockHash
	}
	return nil
}

// Vote creates a vote for block
func (blk *Block) Vote(priv *PrivateKey) *Vote {
	return NewVote(&core_pb.Vote{
		Block:     blk.data.Hash,
		Signature: priv.Sign(blk.data.Hash).data,
	})
}
