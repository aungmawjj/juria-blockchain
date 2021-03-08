// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package merkle

import (
	"math/big"
	"sync"
)

// Store is merkle tree store
type Store interface {
	GetLeafCount() *big.Int
	GetHeight() uint8
	GetNode(p *Position) []byte
}

// MapStore is simple Store implementation
type MapStore struct {
	nleaf  *big.Int
	height uint8
	nodes  map[string][]byte
	mtx    sync.RWMutex
}

var _ Store = (*MapStore)(nil)

// NewMapStore create a new MapStore
func NewMapStore() *MapStore {
	return &MapStore{
		nleaf: big.NewInt(0),
		nodes: make(map[string][]byte),
	}
}

// GetLeafCount godoc
func (ms *MapStore) GetLeafCount() *big.Int {
	ms.mtx.RLock()
	defer ms.mtx.RUnlock()

	return ms.nleaf
}

// GetHeight godoc
func (ms *MapStore) GetHeight() uint8 {
	ms.mtx.RLock()
	defer ms.mtx.RUnlock()

	return ms.height
}

// GetNode godoc
func (ms *MapStore) GetNode(p *Position) []byte {
	ms.mtx.RLock()
	defer ms.mtx.RUnlock()

	return ms.nodes[p.String()]
}

// CommitUpdate godoc
func (ms *MapStore) CommitUpdate(res *UpdateResult) {
	ms.mtx.Lock()
	defer ms.mtx.Unlock()

	ms.nleaf = res.LeafCount
	ms.height = res.Height
	for _, n := range res.Nodes {
		ms.nodes[n.Position.String()] = n.Data
	}
}
