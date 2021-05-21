// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package execution

import (
	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/execution/chaincode"
)

type callContext struct {
	blk   *core.Block
	tx    *core.Transaction
	input []byte
	State
}

var _ chaincode.CallContext = (*callContext)(nil)

func (ctx *callContext) Sender() []byte {
	if ctx.tx == nil {
		return nil
	}
	if ctx.tx.Sender() == nil {
		return nil
	}
	return ctx.tx.Sender().Bytes()
}

func (ctx *callContext) BlockHash() []byte {
	if ctx.blk == nil {
		return nil
	}
	return ctx.blk.Hash()
}

func (ctx *callContext) BlockHeight() uint64 {
	if ctx.blk == nil {
		return 0
	}
	return ctx.blk.Height()
}

func (ctx *callContext) Input() []byte {
	return ctx.input
}
