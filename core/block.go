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

func NewBlock() *Block {
	return &Block{
		data: new(core_pb.Block),
	}
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
	if blk.data == nil {
		return ErrNilBlock
	}
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
	return NewVote().setData(&core_pb.Vote{
		BlockHash: blk.data.Hash,
		Signature: priv.Sign(blk.data.Hash).data,
	})
}

func (blk *Block) setData(data *core_pb.Block) *Block {
	blk.data = data
	blk.quorumCert = NewQuorumCert().setData(data.QuorumCert)
	return blk
}

func (blk *Block) SetHash(val []byte) *Block {
	blk.data.Hash = val
	return blk
}

func (blk *Block) SetHeight(val uint64) *Block {
	blk.data.Height = val
	return blk
}

func (blk *Block) SetParentHash(val []byte) *Block {
	blk.data.ParentHash = val
	return blk
}

func (blk *Block) SetProposer(val []byte) *Block {
	blk.data.Proposer = val
	return blk
}

func (blk *Block) SetQuorumCert(val *QuorumCert) *Block {
	blk.quorumCert = val
	blk.data.QuorumCert = val.data
	return blk
}

func (blk *Block) SetExecHeight(val uint64) *Block {
	blk.data.ExecHeight = val
	return blk
}

func (blk *Block) SetStateRoot(val []byte) *Block {
	blk.data.StateRoot = val
	return blk
}

func (blk *Block) SetTransactions(val [][]byte) *Block {
	blk.data.Transactions = val
	return blk
}

func (blk *Block) Hash() []byte            { return blk.data.Hash }
func (blk *Block) Height() uint64          { return blk.data.Height }
func (blk *Block) ParentHash() []byte      { return blk.data.ParentHash }
func (blk *Block) Proposer() []byte        { return blk.data.Proposer }
func (blk *Block) QuorumCert() *QuorumCert { return blk.quorumCert }
func (blk *Block) ExecHeight() uint64      { return blk.data.ExecHeight }
func (blk *Block) StateRoot() []byte       { return blk.data.StateRoot }
func (blk *Block) Transactions() [][]byte  { return blk.data.Transactions }

// Marshal encodes blk as bytes
func (blk *Block) Marshal() ([]byte, error) {
	return proto.Marshal(blk.data)
}

// UnmarshalBlock decodes block from bytes
func UnmarshalBlock(b []byte) (*Block, error) {
	data := new(core_pb.Block)
	if err := proto.Unmarshal(b, data); err != nil {
		return nil, err
	}
	return NewBlock().setData(data), nil
}
