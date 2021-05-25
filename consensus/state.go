// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package consensus

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/aungmawjj/juria-blockchain/core"
)

type state struct {
	resources *Resources

	blockpool map[string]*core.Block
	mtxBlocks sync.RWMutex

	mtxUpdate sync.Mutex

	leaderIndex int64
}

func newState(resources *Resources) *state {
	return &state{
		resources: resources,
		blockpool: make(map[string]*core.Block),
	}
}

func (state *state) getBlockPoolSize() int {
	state.mtxBlocks.RLock()
	defer state.mtxBlocks.RUnlock()
	return len(state.blockpool)
}

func (state *state) setBlock(blk *core.Block) {
	state.mtxBlocks.Lock()
	defer state.mtxBlocks.Unlock()
	state.blockpool[string(blk.Hash())] = blk
}

func (state *state) getBlock(hash []byte) *core.Block {
	state.mtxBlocks.RLock()
	defer state.mtxBlocks.RUnlock()
	return state.blockpool[string(hash)]
}

func (state *state) deleteBlock(hash []byte) {
	state.mtxBlocks.Lock()
	defer state.mtxBlocks.Unlock()
	delete(state.blockpool, string(hash))
}

func (state *state) getOlderBlocks(blk *core.Block) []*core.Block {
	state.mtxBlocks.RLock()
	defer state.mtxBlocks.RUnlock()

	ret := make([]*core.Block, 0)
	for _, b := range state.blockpool {
		if b.Height() < blk.Height() {
			ret = append(ret, b)
		}
	}
	return ret
}

func (state *state) getBlockOnLocalNode(hash []byte) *core.Block {
	blk := state.getBlock(hash)
	if blk == nil {
		fmt.Println("block not found in state")
		fmt.Println(hash)
		blk, _ = state.resources.Storage.GetBlock(hash)
	}
	return blk
}

func (state *state) isThisNodeLeader() bool {
	return state.isLeader(state.resources.Signer.PublicKey())
}

func (state *state) isLeader(pubKey *core.PublicKey) bool {
	vIdx, ok := state.resources.VldStore.GetValidatorIndex(pubKey)
	if !ok {
		return false
	}
	return state.getLeaderIndex() == vIdx
}

func (state *state) setLeaderIndex(idx int) {
	atomic.StoreInt64(&state.leaderIndex, int64(idx))
}

func (state *state) getLeaderIndex() int {
	return int(atomic.LoadInt64(&state.leaderIndex))
}
