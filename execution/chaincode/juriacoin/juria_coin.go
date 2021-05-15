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

var _ chaincode.ChainCode = (*JuriaCoin)(nil)

func (jrc *JuriaCoin) Init(wc chaincode.WriteContext) error {
	wc.SetState(keyMinter, wc.Sender())
	return nil
}

func (jrc *JuriaCoin) Invoke(wc chaincode.WriteContext) error {
	input, err := parseInput(wc.Input())
	if err != nil {
		return err
	}
	switch input.Method {

	case "setMinter":
		return invokeSetMinter(wc, input)

	case "mint":
		return invokeMint(wc, input)

	case "transfer":
		return invokeTransfer(wc, input)

	default:
		return errors.New("method not found")
	}
}

func (jrc *JuriaCoin) Query(rc chaincode.ReadContext) ([]byte, error) {
	input, err := parseInput(rc.Input())
	if err != nil {
		return nil, err
	}
	switch input.Method {

	case "minter":
		return rc.GetState(keyMinter)

	case "total":
		return queryTotal(rc)

	case "balance":
		return queryBalance(rc, input)

	default:
		return nil, errors.New("method not found")
	}
}

func invokeSetMinter(wc chaincode.WriteContext, input *Input) error {
	minter := wc.GetState(keyMinter)
	if !bytes.Equal(minter, wc.Sender()) {
		return errors.New("sender must be minter")
	}
	wc.SetState(keyMinter, input.Dest)
	return nil
}

func invokeMint(wc chaincode.WriteContext, input *Input) error {
	minter := wc.GetState(keyMinter)
	if !bytes.Equal(minter, wc.Sender()) {
		return errors.New("sender must be minter")
	}
	total := decodeBalance(wc.GetState(keyTotal))
	balance := decodeBalance(wc.GetState(input.Dest))

	total += input.Value
	balance += input.Value

	wc.SetState(keyTotal, encodeBalance(total))
	wc.SetState(input.Dest, encodeBalance(balance))
	return nil
}

func invokeTransfer(wc chaincode.WriteContext, input *Input) error {
	bsrc := decodeBalance(wc.GetState(wc.Sender()))
	if bsrc < input.Value {
		return errors.New("not enough balance")
	}
	bdes := decodeBalance(wc.GetState(input.Dest))

	bsrc -= input.Value
	bdes += input.Value

	wc.SetState(wc.Sender(), encodeBalance(bsrc))
	wc.SetState(input.Dest, encodeBalance(bdes))
	return nil
}

func queryTotal(rc chaincode.ReadContext) ([]byte, error) {
	b, err := rc.GetState(keyTotal)
	if err != nil {
		return nil, err
	}
	return json.Marshal(decodeBalance(b))
}

func queryBalance(rc chaincode.ReadContext, input *Input) ([]byte, error) {
	b, err := rc.GetState(input.Dest)
	if err != nil {
		return nil, err
	}
	return json.Marshal(decodeBalance(b))
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
