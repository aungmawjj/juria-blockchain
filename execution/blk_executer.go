// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package execution

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/emitter"
	"github.com/aungmawjj/juria-blockchain/logger"
)

type blkExecutor struct {
	txTimeout       time.Duration
	concurrentLimit int

	codeRegistry *codeRegistry
	state        StateRO
	blk          *core.Block
	txs          []*core.Transaction

	rootTrk   *stateTracker
	txCommits []*core.TxCommit

	mergeIdx     int32
	mergeEmitter *emitter.Emitter
}

/*
execute transactions of a block in sequential
to improve the performance, execute transactions in parallel
if state conflict occur, (i.e, a transaction call getState of the another transaction's setState)
re-execute the conflict transactions
*/
func (bexe *blkExecutor) execute() (*core.BlockCommit, []*core.TxCommit) {
	start := time.Now()
	bexe.mergeEmitter = emitter.New()
	bexe.rootTrk = newStateTracker(bexe.state, nil)
	bexe.txCommits = make([]*core.TxCommit, len(bexe.txs))
	bexe.executeConcurrent()
	elapsed := time.Since(start)
	bcm := core.NewBlockCommit().
		SetHash(bexe.blk.Hash()).
		SetStateChanges(bexe.rootTrk.getStateChanges()).
		SetElapsedExec(elapsed.Seconds())

	if len(bexe.txs) > 0 {
		logger.I().Debugw("batch execution",
			"txs", len(bexe.txs), "elapsed", elapsed)
	}
	return bcm, bexe.txCommits
}

func (bexe *blkExecutor) executeConcurrent() {
	var wg sync.WaitGroup
	for i := range bexe.txs {
		wg.Add(1)
		go func(i int) {
			bexe.executeTxAndMerge(i)
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func (bexe *blkExecutor) executeTxAndMerge(i int) {
	sub := bexe.mergeEmitter.Subscribe(100)
	defer sub.Unsubscribe()

	// limit the number of concurrent execution
	if i > bexe.concurrentLimit {
		for e := range sub.Events() {
			mergeIdx := e.(int)
			if i-mergeIdx <= bexe.concurrentLimit {
				break
			}
		}
	}

	texe := bexe.executeTx(i)
	if bexe.getMergeIdx() == i {
		bexe.mergeTxStateChanges(i, texe)
		return
	}
	for e := range sub.Events() {
		mergeIdx := e.(int)
		if mergeIdx == i {
			bexe.mergeTxStateChanges(i, texe)
			return
		}
	}
}

func (bexe *blkExecutor) mergeTxStateChanges(i int, texe *txExecutor) {
	if bexe.rootTrk.hasDependencyChanges(texe.rootTrk) {
		// earlier txs changes the dependencies of this tx, execute tx again
		texe = bexe.executeTx(i)
	}
	if bexe.txCommits[i].Error() != "" {
		return // don't merge
	}
	bexe.rootTrk.merge(texe.rootTrk)
	bexe.increaseMergeIdx()
}

func (bexe *blkExecutor) executeTx(i int) *txExecutor {
	texe := &txExecutor{
		codeRegistry: bexe.codeRegistry,
		timeout:      bexe.txTimeout,
		rootTrk:      bexe.rootTrk.spawn(nil),
		blk:          bexe.blk,
		tx:           bexe.txs[i],
	}
	bexe.txCommits[i] = texe.execute()
	return texe
}

func (bexe *blkExecutor) getMergeIdx() int {
	return int(atomic.LoadInt32(&bexe.mergeIdx))
}

func (bexe *blkExecutor) increaseMergeIdx() {
	i := atomic.AddInt32(&bexe.mergeIdx, 1)
	bexe.mergeEmitter.Emit(int(i))
}
