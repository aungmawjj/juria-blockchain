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
	ErrNilBlock         = errors.New("nil block")
)

// Block type
type Block struct {
	data       *core_pb.Block
	quorumCert *QuorumCert
}

// newBlock creates Block from pb data
func newBlock(data *core_pb.Block) (*Block, error) {
	if data == nil {
		return nil, ErrNilBlock
	}
	qc, err := newQuorumCert(data.QuorumCert)
	if err != nil {
		return nil, err
	}
	return &Block{data, qc}, nil
}

// UnmarshalBlock decodes block from bytes
func UnmarshalBlock(b []byte) (*Block, error) {
	data := new(core_pb.Block)
	if err := proto.Unmarshal(b, data); err != nil {
		return nil, err
	}
	return newBlock(data)
}

// Marshal encodes blk as bytes
func (blk *Block) Marshal() ([]byte, error) {
	return proto.Marshal(blk.data)
}

// Sum returns sha3 sum of block
func (blk *Block) Sum() []byte {
	h := sha3.New256()
	h.Write(uint64ToBytes(blk.data.Height))
	h.Write(blk.data.ParentHash)
	h.Write(blk.data.Proposer)
	if blk.data.QuorumCert != nil {
		h.Write(blk.data.QuorumCert.BlockHash) // qc reference block hash
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
	vote, _ := newVote(&core_pb.Vote{
		BlockHash: blk.data.Hash,
		Signature: priv.Sign(blk.data.Hash).data,
	})
	return vote
}

func (blk *Block) setData() *Block {
	if blk.data == nil {
		blk.data = new(core_pb.Block)
	}
	return blk
}

func (blk *Block) SetHash(val []byte) *Block {
	blk.setData()
	blk.data.Hash = val
	return blk
}

func (blk *Block) SetHeight(val uint64) *Block {
	blk.setData()
	blk.data.Height = val
	return blk
}

func (blk *Block) SetParentHash(val []byte) *Block {
	blk.setData()
	blk.data.ParentHash = val
	return blk
}

func (blk *Block) SetProposer(val []byte) *Block {
	blk.setData()
	blk.data.Proposer = val
	return blk
}

func (blk *Block) SetQuorumCert(val *QuorumCert) *Block {
	blk.setData()
	blk.quorumCert = val
	blk.data.QuorumCert = val.data
	return blk
}

func (blk *Block) SetExecHeight(val uint64) *Block {
	blk.setData()
	blk.data.ExecHeight = val
	return blk
}

func (blk *Block) SetStateRoot(val []byte) *Block {
	blk.setData()
	blk.data.StateRoot = val
	return blk
}

func (blk *Block) SetTransactions(val [][]byte) *Block {
	blk.setData()
	blk.data.Transactions = val
	return blk
}
