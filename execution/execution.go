// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package execution

import (
	"encoding/json"
	"time"

	"github.com/aungmawjj/juria-blockchain/core"
)

type Config struct {
	TxExecTimeout time.Duration
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

func (exec *Execution) Query(codeAddr, input []byte) ([]byte, error) {
	cc, err := exec.codeRegistry.getInstance(
		codeAddr, newStateTracker(exec.state, codeRegistryAddr))
	if err != nil {
		return nil, err
	}
	return cc.Query(&callContext{
		input: input,
		State: newStateTracker(exec.state, codeAddr),
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

func (bexe *blkExecutor) execute() (*core.BlockCommit, []*core.TxCommit) {
	start := time.Now()
	bexe.rootTrk = newStateTracker(bexe.state, nil)
	bexe.txCommits = make([]*core.TxCommit, len(bexe.txs))
	for i := range bexe.txs {
		bexe.executeTx(i)
	}
	bcm := core.NewBlockCommit().
		SetHash(bexe.blk.Hash()).
		SetStateChanges(bexe.rootTrk.getStateChanges()).
		SetElapsedExec(time.Since(start).Seconds())

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
