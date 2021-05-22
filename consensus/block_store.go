// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package consensus

import (
	"sync"

	"github.com/aungmawjj/juria-blockchain/core"
)

type blockStore struct {
	blocks map[string]*core.Block
	mtx    sync.RWMutex
}

func newBlockStore() *blockStore {
	return &blockStore{
		blocks: make(map[string]*core.Block),
	}
}

func (store *blockStore) setBlock(blk *core.Block) {
	store.mtx.Lock()
	defer store.mtx.Unlock()
	store.blocks[string(blk.Hash())] = blk
}

func (store *blockStore) getBlock(hash []byte) *core.Block {
	store.mtx.RLock()
	defer store.mtx.RUnlock()
	return store.blocks[string(hash)]
}

func (store *blockStore) deleteBlock(hash []byte) {
	store.mtx.Lock()
	defer store.mtx.Unlock()
	delete(store.blocks, string(hash))
}

func (store *blockStore) getOlderBlocks(blk *core.Block) []*core.Block {
	store.mtx.RLock()
	defer store.mtx.RUnlock()

	ret := make([]*core.Block, 0)
	for _, b := range store.blocks {
		if b.Height() < blk.Height() {
			ret = append(ret, b)
		}
	}
	return ret
}
