// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package core

import (
	"github.com/aungmawjj/juria-blockchain/core/core_pb"
	"google.golang.org/protobuf/proto"
)

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

func (bcm *BlockCommit) SetOldBlockTxs(val [][]byte) *BlockCommit {
	bcm.data.OldBlockTxs = val
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
func (bcm *BlockCommit) OldBlockTxs() [][]byte  { return bcm.data.OldBlockTxs }
func (bcm *BlockCommit) LeafCount() []byte      { return bcm.data.LeafCount }
func (bcm *BlockCommit) MerkleRoot() []byte     { return bcm.data.MerkleRoot }
func (bcm *BlockCommit) ElapsedExec() float64   { return bcm.data.ElapsedExec }
func (bcm *BlockCommit) ElapsedMerkle() float64 { return bcm.data.ElapsedMerkle }

func (bcm *BlockCommit) StateChanges() []*StateChange {
	scList := make([]*StateChange, len(bcm.data.StateChanges))
	for i, scData := range bcm.data.StateChanges {
		sc := NewStateChange()
		sc.setData(scData)
		scList[i] = sc
	}
	return scList
}

func (bcm *BlockCommit) setData(data *core_pb.BlockCommit) error {
	bcm.data = data
	return nil
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
	return bcm.setData(data)
}
