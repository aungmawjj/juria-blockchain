// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package execution

import (
	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/execution/chaincode"
)

type callContextTx struct {
	blk   *core.Block
	tx    *core.Transaction
	input []byte
	*stateTracker
}

var _ chaincode.CallContext = (*callContextTx)(nil)

func (ctx *callContextTx) Sender() []byte {
	if ctx.tx == nil {
		return nil
	}
	if ctx.tx.Sender() == nil {
		return nil
	}
	return ctx.tx.Sender().Bytes()
}

func (ctx *callContextTx) BlockHash() []byte {
	if ctx.blk == nil {
		return nil
	}
	return ctx.blk.Hash()
}

func (ctx *callContextTx) BlockHeight() uint64 {
	if ctx.blk == nil {
		return 0
	}
	return ctx.blk.Height()
}

func (ctx *callContextTx) Input() []byte {
	return ctx.input
}

type callContextQuery struct {
	input []byte
	stateGetter
}

var _ chaincode.CallContext = (*callContextQuery)(nil)

func (ctx *callContextQuery) Input() []byte {
	return ctx.input
}

func (ctx *callContextQuery) Sender() []byte {
	return nil
}

func (ctx *callContextQuery) BlockHash() []byte {
	return nil
}

func (ctx *callContextQuery) BlockHeight() uint64 {
	return 0
}

func (ctx *callContextQuery) SetState(key, value []byte) {
	// do nothing
}
