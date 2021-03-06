// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package consensus

import (
	"sync"
	"sync/atomic"

	"github.com/aungmawjj/juria-blockchain/core"
)

type state struct {
	resources *Resources

	blocks    map[string]*core.Block
	mtxBlocks sync.RWMutex

	commited    map[string]struct{}
	mtxCommited sync.RWMutex

	qcs    map[string]*core.QuorumCert // qc by block hash
	mtxQCs sync.RWMutex

	mtxUpdate sync.Mutex // lock for hotstuff update call

	leaderIndex int64

	// commited block height. on node restart, it's zero until a block is commited
	commitedHeight uint64

	// commited tx count, since last node start
	commitedTxCount uint64
}

func newState(resources *Resources) *state {
	return &state{
		resources: resources,
		blocks:    make(map[string]*core.Block),
		commited:  make(map[string]struct{}),
		qcs:       make(map[string]*core.QuorumCert),
	}
}

func (state *state) getBlockPoolSize() int {
	state.mtxBlocks.RLock()
	defer state.mtxBlocks.RUnlock()
	return len(state.blocks)
}

func (state *state) setBlock(blk *core.Block) {
	state.mtxBlocks.Lock()
	defer state.mtxBlocks.Unlock()
	state.blocks[string(blk.Hash())] = blk
}

func (state *state) getBlock(hash []byte) *core.Block {
	blk := state.getBlockFromState(hash)
	if blk != nil {
		return blk
	}
	blk, _ = state.resources.Storage.GetBlock(hash)
	if blk == nil {
		return nil
	}
	state.setBlock(blk)
	return blk
}

func (state *state) getBlockFromState(hash []byte) *core.Block {
	state.mtxBlocks.RLock()
	defer state.mtxBlocks.RUnlock()
	return state.blocks[string(hash)]
}

func (state *state) deleteBlock(hash []byte) {
	state.mtxBlocks.Lock()
	defer state.mtxBlocks.Unlock()
	delete(state.blocks, string(hash))
}

func (state *state) getQCPoolSize() int {
	state.mtxQCs.RLock()
	defer state.mtxQCs.RUnlock()
	return len(state.qcs)
}

func (state *state) setQC(qc *core.QuorumCert) {
	state.mtxQCs.Lock()
	defer state.mtxQCs.Unlock()
	state.qcs[string(qc.BlockHash())] = qc
}

func (state *state) getQC(blkHash []byte) *core.QuorumCert {
	state.mtxQCs.RLock()
	defer state.mtxQCs.RUnlock()
	return state.qcs[string(blkHash)]
}

func (state *state) deleteQC(blkHash []byte) {
	state.mtxQCs.Lock()
	defer state.mtxQCs.Unlock()
	delete(state.qcs, string(blkHash))
}

func (state *state) setCommitedBlock(blk *core.Block) {
	state.mtxCommited.Lock()
	defer state.mtxCommited.Unlock()
	state.commited[string(blk.Hash())] = struct{}{}
	atomic.StoreUint64(&state.commitedHeight, blk.Height())
}

func (state *state) deleteCommited(blkhash []byte) {
	state.mtxCommited.Lock()
	defer state.mtxCommited.Unlock()
	delete(state.commited, string(blkhash))
}

func (state *state) getOlderBlocks(height uint64) []*core.Block {
	state.mtxBlocks.RLock()
	defer state.mtxBlocks.RUnlock()
	ret := make([]*core.Block, 0)
	for _, b := range state.blocks {
		if b.Height() < height {
			ret = append(ret, b)
		}
	}
	return ret
}

func (state *state) getUncommitedOlderBlocks(bexec *core.Block) []*core.Block {
	state.mtxBlocks.RLock()
	defer state.mtxBlocks.RUnlock()

	state.mtxCommited.RLock()
	defer state.mtxCommited.RUnlock()

	ret := make([]*core.Block, 0)
	for _, b := range state.blocks {
		if b.Height() < bexec.Height() {
			if _, commited := state.commited[string(b.Hash())]; !commited {
				ret = append(ret, b)
			}
		}
	}
	return ret
}

func (state *state) isThisNodeLeader() bool {
	return state.isLeader(state.resources.Signer.PublicKey())
}

func (state *state) isLeader(pubKey *core.PublicKey) bool {
	if !state.resources.VldStore.IsValidator(pubKey) {
		return false
	}
	return state.getLeaderIndex() == state.resources.VldStore.GetValidatorIndex(pubKey)
}

func (state *state) setLeaderIndex(idx int) {
	atomic.StoreInt64(&state.leaderIndex, int64(idx))
}

func (state *state) getLeaderIndex() int {
	return int(atomic.LoadInt64(&state.leaderIndex))
}

func (state *state) getFaultyCount() int {
	return state.resources.VldStore.ValidatorCount() - state.resources.VldStore.MajorityCount()
}

func (state *state) addCommitedTxCount(count int) {
	atomic.AddUint64(&state.commitedTxCount, uint64(count))
}

func (state *state) getCommitedTxCount() int {
	return int(atomic.LoadUint64(&state.commitedTxCount))
}

func (state *state) getCommitedHeight() uint64 {
	return atomic.LoadUint64(&state.commitedHeight)
}
