// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package execution

import (
	"encoding/json"
	"time"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/execution/bincc"
	"github.com/aungmawjj/juria-blockchain/logger"
)

type Config struct {
	TxExecTimeout time.Duration
	BinccDir      string
}

var DefaultConfig = Config{
	TxExecTimeout: 10 * time.Second,
}

type Execution struct {
	state  StateRO
	config Config

	codeRegistry *codeRegistry
}

func New(state StateRO, config Config) *Execution {
	exec := &Execution{
		state:  state,
		config: config,
	}
	exec.codeRegistry = newCodeRegistry()
	exec.codeRegistry.registerDriver(DriverTypeNative, newNativeCodeDriver())
	exec.codeRegistry.registerDriver(DriverTypeBincc,
		bincc.NewCodeDriver(exec.config.BinccDir, exec.config.TxExecTimeout))
	return exec
}

func (exec *Execution) Execute(blk *core.Block, txs []*core.Transaction) (
	*core.BlockCommit, []*core.TxCommit,
) {
	bexe := &blkExecutor{
		codeRegistry: exec.codeRegistry,
		state:        exec.state,
		txTimeout:    exec.config.TxExecTimeout,
		blk:          blk,
		txs:          txs,
	}
	return bexe.execute()
}

type QueryData struct {
	CodeAddr []byte
	Input    []byte
}

func (exec *Execution) Query(query *QueryData) ([]byte, error) {
	cc, err := exec.codeRegistry.getInstance(
		query.CodeAddr, newStateTracker(exec.state, codeRegistryAddr))
	if err != nil {
		return nil, err
	}
	return cc.Query(&callContext{
		input: query.Input,
		State: newStateTracker(exec.state, query.CodeAddr),
	})
}

func (exec *Execution) VerifyTx(tx *core.Transaction) error {
	if len(tx.CodeAddr()) != 0 { // invoke tx
		return nil
	}
	// deployment tx
	input := new(DeploymentInput)
	err := json.Unmarshal(tx.Input(), input)
	if err != nil {
		return err
	}
	return exec.codeRegistry.install(input)
}

type blkExecutor struct {
	codeRegistry *codeRegistry
	txTimeout    time.Duration
	state        StateRO
	blk          *core.Block
	txs          []*core.Transaction

	rootTrk   *stateTracker
	txCommits []*core.TxCommit
}

/*
execute transactions of a block in sequential
to improve the performance, execute transactions in parallel
if state conflict occur, (i.e, a transaction call getState of the another transaction's setState)
re-execute the conflict transactions
*/
func (bexe *blkExecutor) execute() (*core.BlockCommit, []*core.TxCommit) {
	start := time.Now()
	bexe.rootTrk = newStateTracker(bexe.state, nil)
	bexe.txCommits = make([]*core.TxCommit, len(bexe.txs))
	for i := range bexe.txs {
		bexe.executeTx(i)
	}
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

func (bexe *blkExecutor) executeTx(i int) {
	texe := &txExecutor{
		codeRegistry: bexe.codeRegistry,
		timeout:      bexe.txTimeout,
		rootTrk:      bexe.rootTrk.spawn(nil),
		blk:          bexe.blk,
		tx:           bexe.txs[i],
	}
	bexe.txCommits[i] = texe.execute()
	if bexe.txCommits[i].Error() == "" {
		// if tx is executed without error, merge the state changes
		bexe.rootTrk.merge(texe.rootTrk)
	}
}
