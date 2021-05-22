// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package core

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/aungmawjj/juria-blockchain/core/core_pb"
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
	proposer   *PublicKey
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
	binary.Write(h, binary.BigEndian, blk.data.Height)
	h.Write(blk.data.ParentHash)
	h.Write(blk.data.Proposer)
	if blk.data.QuorumCert != nil {
		h.Write(blk.data.QuorumCert.BlockHash) // qc reference block hash
	}
	binary.Write(h, binary.BigEndian, blk.data.ExecHeight)
	h.Write(blk.data.MerkleRoot)
	binary.Write(h, binary.BigEndian, blk.data.Timestamp)
	for _, txHash := range blk.data.Transactions {
		h.Write(txHash)
	}
	return h.Sum(nil)
}

// Validate block
func (blk *Block) Validate(vs ValidatorStore) error {
	if blk.data == nil {
		return ErrNilBlock
	}
	if err := blk.quorumCert.Validate(vs); err != nil {
		return err
	}
	if !bytes.Equal(blk.Sum(), blk.Hash()) {
		return ErrInvalidBlockHash
	}
	sig, err := newSignature(&core_pb.Signature{
		PubKey: blk.data.Proposer,
		Value:  blk.data.Signature,
	})
	if !vs.IsValidator(sig.PublicKey()) {
		return ErrInvalidValidator
	}
	if err != nil {
		return err
	}
	if !sig.Verify(blk.data.Hash) {
		return ErrInvalidSig
	}
	return nil
}

// Vote creates a vote for block
func (blk *Block) Vote(signer Signer) *Vote {
	return &Vote{
		data: &core_pb.Vote{
			BlockHash: blk.data.Hash,
			Signature: signer.Sign(blk.data.Hash).data,
		},
	}
}

func (blk *Block) setData(data *core_pb.Block) *Block {
	blk.data = data
	blk.quorumCert = NewQuorumCert()
	blk.quorumCert.setData(data.QuorumCert)
	blk.proposer, _ = NewPublicKey(blk.data.Proposer)
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

func (blk *Block) SetQuorumCert(val *QuorumCert) *Block {
	blk.quorumCert = val
	blk.data.QuorumCert = val.data
	return blk
}

func (blk *Block) SetExecHeight(val uint64) *Block {
	blk.data.ExecHeight = val
	return blk
}

func (blk *Block) SetMerkleRoot(val []byte) *Block {
	blk.data.MerkleRoot = val
	return blk
}

func (blk *Block) SetTimestamp(val int64) *Block {
	blk.data.Timestamp = val
	return blk
}

func (blk *Block) SetTransactions(val [][]byte) *Block {
	blk.data.Transactions = val
	return blk
}

func (blk *Block) Sign(signer Signer) *Block {
	blk.proposer = signer.PublicKey()
	blk.data.Proposer = signer.PublicKey().key
	blk.data.Hash = blk.Sum()
	blk.data.Signature = signer.Sign(blk.data.Hash).data.Value
	return blk
}

func (blk *Block) Hash() []byte            { return blk.data.Hash }
func (blk *Block) Height() uint64          { return blk.data.Height }
func (blk *Block) ParentHash() []byte      { return blk.data.ParentHash }
func (blk *Block) Proposer() *PublicKey    { return blk.proposer }
func (blk *Block) QuorumCert() *QuorumCert { return blk.quorumCert }
func (blk *Block) ExecHeight() uint64      { return blk.data.ExecHeight }
func (blk *Block) MerkleRoot() []byte      { return blk.data.MerkleRoot }
func (blk *Block) Timestamp() int64        { return blk.data.Timestamp }
func (blk *Block) Transactions() [][]byte  { return blk.data.Transactions }

// Marshal encodes blk as bytes
func (blk *Block) Marshal() ([]byte, error) {
	return proto.Marshal(blk.data)
}

// UnmarshalBlock decodes block from bytes
func (blk *Block) Unmarshal(b []byte) error {
	data := new(core_pb.Block)
	if err := proto.Unmarshal(b, data); err != nil {
		return err
	}
	blk.setData(data)
	return nil
}

type BlockCommit struct {
	data *core_pb.BlockCommit
}

func NewBlockCommit() *BlockCommit {
	return &BlockCommit{
		data: new(core_pb.BlockCommit),
	}
}

func (bcm *BlockCommit) SetHash(val []byte) *BlockCommit {
	bcm.data.Hash = val
	return bcm
}

func (bcm *BlockCommit) SetLeafCount(val []byte) *BlockCommit {
	bcm.data.LeafCount = val
	return bcm
}

func (bcm *BlockCommit) SetMerkleRoot(val []byte) *BlockCommit {
	bcm.data.MerkleRoot = val
	return bcm
}

func (bcm *BlockCommit) SetElapsedExec(val float64) *BlockCommit {
	bcm.data.ElapsedExec = val
	return bcm
}

func (bcm *BlockCommit) SetElapsedMerkle(val float64) *BlockCommit {
	bcm.data.ElapsedMerkle = val
	return bcm
}

func (bcm *BlockCommit) SetStateChanges(val []*StateChange) *BlockCommit {
	scpb := make([]*core_pb.StateChange, len(val))
	for i, sc := range val {
		scpb[i] = sc.data
	}
	bcm.data.StateChanges = scpb
	return bcm
}

func (bcm *BlockCommit) Hash() []byte           { return bcm.data.Hash }
func (bcm *BlockCommit) LeafCount() []byte      { return bcm.data.LeafCount }
func (bcm *BlockCommit) MerkleRoot() []byte     { return bcm.data.MerkleRoot }
func (bcm *BlockCommit) ElapsedExec() float64   { return bcm.data.ElapsedExec }
func (bcm *BlockCommit) ElapsedMerkle() float64 { return bcm.data.ElapsedMerkle }

func (bcm *BlockCommit) StateChanges() []*StateChange {
	scList := make([]*StateChange, len(bcm.data.StateChanges))
	for i, scData := range bcm.data.StateChanges {
		scList[i] = NewStateChange().setData(scData)
	}
	return scList
}

func (bcm *BlockCommit) setData(data *core_pb.BlockCommit) *BlockCommit {
	bcm.data = data
	return bcm
}

func (bcm *BlockCommit) Marshal() ([]byte, error) {
	return proto.Marshal(bcm.data)
}

func (bcm *BlockCommit) Unmarshal(b []byte) error {
	data := new(core_pb.BlockCommit)
	err := proto.Unmarshal(b, data)
	if err != nil {
		return err
	}
	bcm.setData(data)
	return nil
}
