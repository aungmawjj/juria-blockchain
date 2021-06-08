// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package juriacoin

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"

	"github.com/aungmawjj/juria-blockchain/execution/chaincode"
)

type Input struct {
	Method string `json:"method"`
	Dest   []byte `json:"dest"`
	Value  int64  `json:"value"`
}

var (
	keyMinter = []byte("minter")
	keyTotal  = []byte("total")
)

// JuriaCoin chaincode
type JuriaCoin struct{}

var _ chaincode.Chaincode = (*JuriaCoin)(nil)

func (jctx *JuriaCoin) Init(ctx chaincode.CallContext) error {
	ctx.SetState(keyMinter, ctx.Sender())
	return nil
}

func (jctx *JuriaCoin) Invoke(ctx chaincode.CallContext) error {
	input, err := parseInput(ctx.Input())
	if err != nil {
		return err
	}
	switch input.Method {

	case "setMinter":
		return invokeSetMinter(ctx, input)

	case "mint":
		return invokeMint(ctx, input)

	case "transfer":
		return invokeTransfer(ctx, input)

	default:
		return errors.New("method not found")
	}
}

func (jctx *JuriaCoin) Query(ctx chaincode.CallContext) ([]byte, error) {
	input, err := parseInput(ctx.Input())
	if err != nil {
		return nil, err
	}
	switch input.Method {

	case "minter":
		return ctx.GetState(keyMinter), nil

	case "total":
		return queryTotal(ctx)

	case "balance":
		return queryBalance(ctx, input)

	default:
		return nil, errors.New("method not found")
	}
}

func invokeSetMinter(ctx chaincode.CallContext, input *Input) error {
	minter := ctx.GetState(keyMinter)
	if !bytes.Equal(minter, ctx.Sender()) {
		return errors.New("sender must be minter")
	}
	ctx.SetState(keyMinter, input.Dest)
	return nil
}

func invokeMint(ctx chaincode.CallContext, input *Input) error {
	minter := ctx.GetState(keyMinter)
	if !bytes.Equal(minter, ctx.Sender()) {
		return errors.New("sender must be minter")
	}
	total := decodeBalance(ctx.GetState(keyTotal))
	balance := decodeBalance(ctx.GetState(input.Dest))

	total += input.Value
	balance += input.Value

	ctx.SetState(keyTotal, encodeBalance(total))
	ctx.SetState(input.Dest, encodeBalance(balance))
	return nil
}

func invokeTransfer(ctx chaincode.CallContext, input *Input) error {
	bsctx := decodeBalance(ctx.GetState(ctx.Sender()))
	if bsctx < input.Value {
		return errors.New("not enough balance")
	}
	bdes := decodeBalance(ctx.GetState(input.Dest))

	bsctx -= input.Value
	bdes += input.Value

	ctx.SetState(ctx.Sender(), encodeBalance(bsctx))
	ctx.SetState(input.Dest, encodeBalance(bdes))
	return nil
}

func queryTotal(ctx chaincode.CallContext) ([]byte, error) {
	return json.Marshal(decodeBalance(ctx.GetState(keyTotal)))
}

func queryBalance(ctx chaincode.CallContext, input *Input) ([]byte, error) {
	return json.Marshal(decodeBalance(ctx.GetState(input.Dest)))
}

func decodeBalance(b []byte) int64 {
	if b == nil {
		return 0
	}
	return int64(binary.BigEndian.Uint64(b))
}

func encodeBalance(value int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(value))
	return b
}

func parseInput(b []byte) (*Input, error) {
	input := new(Input)
	err := json.Unmarshal(b, input)
	if err != nil {
		return nil, errors.New("failed to parse input: " + err.Error())
	}
	return input, nil
}
