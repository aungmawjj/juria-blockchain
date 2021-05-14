// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package storage

import (
	"math/big"

	"github.com/aungmawjj/juria-blockchain/merkle"
)

type merkleStore struct {
	getter getter
}

var _ merkle.Store = (*merkleStore)(nil)

func (ms *merkleStore) GetLeafCount() *big.Int {
	return ms.getLeafCount()
}

func (ms *merkleStore) GetHeight() uint8 {
	return ms.getHeight()
}

func (ms *merkleStore) GetNode(p *merkle.Position) []byte {
	return ms.getNode(p)
}

func (ms *merkleStore) commitUpdate(upd *merkle.UpdateResult) []updateFunc {
	ret := make([]updateFunc, 0)
	ret = append(ret, ms.setNodes(upd.Leaves)...)
	ret = append(ret, ms.setNodes(upd.Branches)...)
	ret = append(ret, ms.setLeafCount(upd.LeafCount))
	ret = append(ret, ms.setTreeHeight(upd.Height))
	return ret
}

func (ms *merkleStore) getNode(p *merkle.Position) []byte {
	val, _ := ms.getter.Get(concatBytes([]byte{colMerkleNodeByPosition}, p.Bytes()))
	return val
}

func (ms *merkleStore) getLeafCount() *big.Int {
	count := big.NewInt(0)
	val, err := ms.getter.Get([]byte{colMerkleLeafCount})
	if err == nil {
		count.SetBytes(val)
	}
	return count
}

func (ms *merkleStore) getHeight() uint8 {
	var height uint8
	val, _ := ms.getter.Get([]byte{colMerkleTreeHeight})
	if len(val) > 0 {
		height = val[0]
	}
	return height
}

func (ms *merkleStore) setNodes(nodes []*merkle.Node) []updateFunc {
	ret := make([]updateFunc, len(nodes))
	for i, n := range nodes {
		ret[i] = ms.setNode(n)
	}
	return ret
}

func (ms *merkleStore) setNode(n *merkle.Node) updateFunc {
	return func(setter setter) error {
		return setter.Set(
			concatBytes([]byte{colMerkleNodeByPosition}, n.Position.Bytes()), n.Data,
		)
	}
}

func (ms *merkleStore) setLeafCount(leafCount *big.Int) updateFunc {
	return func(setter setter) error {
		return setter.Set([]byte{colMerkleLeafCount}, leafCount.Bytes())
	}
}

func (ms *merkleStore) setTreeHeight(height uint8) updateFunc {
	return func(setter setter) error {
		return setter.Set([]byte{colMerkleTreeHeight}, []byte{height})
	}
}
