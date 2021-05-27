// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package execution

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/execution/chaincode"
)

type DeploymentInput struct {
	CodeInfo    CodeInfo `json:"codeInfo"`
	InstallData []byte   `json:"installData"`
	InitInput   []byte   `json:"initInput"`
}

type txExecutor struct {
	codeRegistry *codeRegistry

	timeout time.Duration
	rootTrk *stateTracker

	blk *core.Block
	tx  *core.Transaction
}

func (txe *txExecutor) execute() *core.TxCommit {
	start := time.Now()
	txc := core.NewTxCommit().
		SetHash(txe.tx.Hash()).
		SetBlockHash(txe.blk.Hash()).
		SetBlockHeight(txe.blk.Height())

	err := txe.executeWithTimeout()
	if err != nil {
		txc.SetError(err.Error())
	}
	txc.SetElapsed(time.Since(start).Seconds())
	return txc
}

func (txe *txExecutor) executeWithTimeout() error {
	exeError := make(chan error)
	go func() {
		exeError <- txe.executeChaincode()
	}()

	select {
	case err := <-exeError:
		return err

	case <-time.After(txe.timeout):
		return errors.New("tx execution timeout")
	}
}

func (txe *txExecutor) executeChaincode() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%+v", r)
		}
	}()
	if len(txe.tx.CodeAddr()) == 0 {
		return txe.executeDeployment()
	}
	return txe.executeInvoke()
}

func (txe *txExecutor) executeDeployment() error {
	input := new(DeploymentInput)
	err := json.Unmarshal(txe.tx.Input(), input)
	if err != nil {
		return err
	}

	regTrk := txe.rootTrk.spawn(codeRegistryAddr)
	cc, err := txe.codeRegistry.deploy(txe.tx.Hash(), input, regTrk)
	if err != nil {
		return err
	}

	initTrk := txe.rootTrk.spawn(txe.tx.Hash())
	err = cc.Init(txe.makeCallContext(initTrk, input.InitInput))
	if err != nil {
		return err
	}
	txe.rootTrk.merge(regTrk)
	txe.rootTrk.merge(initTrk)
	return nil
}

func (txe *txExecutor) executeInvoke() error {
	cc, err := txe.codeRegistry.getInstance(txe.tx.CodeAddr(), txe.rootTrk.spawn(codeRegistryAddr))
	if err != nil {
		return err
	}
	invokeTrk := txe.rootTrk.spawn(txe.tx.CodeAddr())
	err = cc.Invoke(txe.makeCallContext(invokeTrk, txe.tx.Input()))
	if err != nil {
		return err
	}
	txe.rootTrk.merge(invokeTrk)
	return nil
}

func (txe *txExecutor) makeCallContext(state State, input []byte) chaincode.CallContext {
	return &callContext{
		blk:   txe.blk,
		tx:    txe.tx,
		input: input,
		State: state,
	}
}
