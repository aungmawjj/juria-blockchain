// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package execution

import (
	"encoding/json"
	"time"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/execution/bincc"
)

type Config struct {
	BinccDir        string
	TxExecTimeout   time.Duration
	ConcurrentLimit int
}

var DefaultConfig = Config{
	TxExecTimeout:   10 * time.Second,
	ConcurrentLimit: 16,
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
		txTimeout:       exec.config.TxExecTimeout,
		concurrentLimit: exec.config.ConcurrentLimit,
		codeRegistry:    exec.codeRegistry,
		state:           exec.state,
		blk:             blk,
		txs:             txs,
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
