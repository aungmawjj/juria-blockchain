// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package juriacoin

import (
	"encoding/json"
	"testing"

	"github.com/aungmawjj/juria-blockchain/execution/chaincode"
	"github.com/stretchr/testify/assert"
)

func TestJuriaCoin_Init(t *testing.T) {
	assert := assert.New(t)
	state := chaincode.NewMockState()
	jctx := new(JuriaCoin)

	ctx := new(chaincode.MockCallContext)
	ctx.MockState = state
	ctx.MockSender = []byte{1, 1, 1}
	err := jctx.Init(ctx)

	assert.NoError(err)

	input := &Input{
		Method: "minter",
	}
	b, _ := json.Marshal(input)
	ctx.MockInput = b
	minter, err := jctx.Query(ctx)

	assert.NoError(err)
	assert.Equal(ctx.MockSender, minter, "deployer should be minter")

	input = &Input{
		Method: "balance",
	}
}

func TestJuriaCoin_SetMinter(t *testing.T) {
	assert := assert.New(t)
	state := chaincode.NewMockState()
	jctx := new(JuriaCoin)

	ctx := new(chaincode.MockCallContext)
	ctx.MockState = state
	ctx.MockSender = []byte{1, 1, 1}
	jctx.Init(ctx)

	input := &Input{
		Method: "setMinter",
		Dest:   []byte{2, 2, 2},
	}
	b, _ := json.Marshal(input)
	ctx.MockSender = []byte{3, 3, 3}
	ctx.MockInput = b
	err := jctx.Invoke(ctx)
	assert.Error(err, "sender not minter error")

	ctx.MockSender = []byte{1, 1, 1}
	err = jctx.Invoke(ctx)

	assert.NoError(err)
	input = &Input{
		Method: "minter",
	}
	b, _ = json.Marshal(input)
	ctx.MockInput = b
	minter, err := jctx.Query(ctx)

	assert.NoError(err)
	assert.Equal([]byte{2, 2, 2}, minter)
}

func TestJuriaCoin_Mint(t *testing.T) {
	assert := assert.New(t)
	state := chaincode.NewMockState()
	jctx := new(JuriaCoin)

	ctx := new(chaincode.MockCallContext)
	ctx.MockState = state
	ctx.MockSender = []byte{1, 1, 1}
	jctx.Init(ctx)

	input := &Input{
		Method: "mint",
		Dest:   []byte{2, 2, 2},
		Value:  100,
	}
	b, _ := json.Marshal(input)
	ctx.MockSender = []byte{3, 3, 3}
	ctx.MockInput = b
	err := jctx.Invoke(ctx)
	assert.Error(err, "sender not minter error")

	ctx.MockSender = []byte{1, 1, 1}
	err = jctx.Invoke(ctx)

	assert.NoError(err)

	input = &Input{
		Method: "total",
	}
	b, _ = json.Marshal(input)
	ctx.MockInput = b
	b, err = jctx.Query(ctx)

	assert.NoError(err)

	var balance int64
	json.Unmarshal(b, &balance)

	assert.EqualValues(100, balance)

	input = &Input{
		Method: "balance",
		Dest:   []byte{2, 2, 2},
	}
	b, _ = json.Marshal(input)
	ctx.MockInput = b
	b, err = jctx.Query(ctx)

	assert.NoError(err)

	balance = 0
	json.Unmarshal(b, &balance)

	assert.EqualValues(100, balance)
}

func TestJuriaCoin_Transfer(t *testing.T) {
	assert := assert.New(t)
	state := chaincode.NewMockState()
	jctx := new(JuriaCoin)

	ctx := new(chaincode.MockCallContext)
	ctx.MockState = state
	ctx.MockSender = []byte{1, 1, 1}
	jctx.Init(ctx)

	input := &Input{
		Method: "mint",
		Dest:   []byte{2, 2, 2},
		Value:  100,
	}
	b, _ := json.Marshal(input)
	ctx.MockInput = b
	jctx.Invoke(ctx)

	// transfer 222 -> 333, value = 101
	input = &Input{
		Method: "transfer",
		Dest:   []byte{3, 3, 3},
		Value:  101,
	}
	b, _ = json.Marshal(input)
	ctx.MockSender = []byte{2, 2, 2}
	ctx.MockInput = b
	err := jctx.Invoke(ctx)

	assert.Error(err, "not enough coin error")

	// transfer 222 -> 333, value = 100
	input.Value = 100
	b, _ = json.Marshal(input)
	ctx.MockInput = b
	err = jctx.Invoke(ctx)

	assert.NoError(err)

	input.Method = "total"
	b, _ = json.Marshal(input)
	ctx.MockInput = b
	b, _ = jctx.Query(ctx)
	var balance int64
	json.Unmarshal(b, &balance)

	assert.EqualValues(100, balance, "total should not change")

	input.Method = "balance"
	input.Dest = []byte{2, 2, 2}
	b, _ = json.Marshal(input)
	ctx.MockInput = b
	b, _ = jctx.Query(ctx)
	balance = 0
	json.Unmarshal(b, &balance)

	assert.EqualValues(0, balance)

	input.Dest = []byte{3, 3, 3}
	b, _ = json.Marshal(input)
	ctx.MockInput = b
	b, _ = jctx.Query(ctx)
	balance = 0
	json.Unmarshal(b, &balance)

	assert.EqualValues(100, balance)
}
