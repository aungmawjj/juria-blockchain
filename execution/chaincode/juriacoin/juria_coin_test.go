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
	jrc := new(JuriaCoin)

	wc := new(chaincode.MockWriteContext)
	wc.MockSender = []byte{1, 1, 1}
	wc.State = state
	err := jrc.Init(wc)

	assert.NoError(err)

	input := &Input{
		Method: "minter",
	}
	b, _ := json.Marshal(input)
	rc := new(chaincode.MockReadContext)
	rc.MockInput = b
	rc.State = state
	minter, err := jrc.Query(rc)

	assert.NoError(err)
	assert.Equal(wc.MockSender, minter, "deployer should be minter")

	input = &Input{
		Method: "balance",
	}
}

func TestJuriaCoin_SetMinter(t *testing.T) {
	assert := assert.New(t)
	state := chaincode.NewMockState()
	jrc := new(JuriaCoin)

	wc := new(chaincode.MockWriteContext)
	wc.MockSender = []byte{1, 1, 1}
	wc.State = state
	jrc.Init(wc)

	input := &Input{
		Method: "setMinter",
		Dest:   []byte{2, 2, 2},
	}
	b, _ := json.Marshal(input)
	wc.MockSender = []byte{3, 3, 3}
	wc.MockInput = b
	err := jrc.Invoke(wc)
	assert.Error(err, "sender not minter error")

	wc.MockSender = []byte{1, 1, 1}
	err = jrc.Invoke(wc)

	assert.NoError(err)
	input = &Input{
		Method: "minter",
	}
	b, _ = json.Marshal(input)
	rc := new(chaincode.MockReadContext)
	rc.State = state
	rc.MockInput = b
	minter, err := jrc.Query(rc)

	assert.NoError(err)
	assert.Equal([]byte{2, 2, 2}, minter)
}

func TestJuriaCoin_Mint(t *testing.T) {
	assert := assert.New(t)
	state := chaincode.NewMockState()
	jrc := new(JuriaCoin)

	wc := new(chaincode.MockWriteContext)
	wc.MockSender = []byte{1, 1, 1}
	wc.State = state
	jrc.Init(wc)

	input := &Input{
		Method: "mint",
		Dest:   []byte{2, 2, 2},
		Value:  100,
	}
	b, _ := json.Marshal(input)
	wc.MockSender = []byte{3, 3, 3}
	wc.MockInput = b
	err := jrc.Invoke(wc)
	assert.Error(err, "sender not minter error")

	wc.MockSender = []byte{1, 1, 1}
	err = jrc.Invoke(wc)

	assert.NoError(err)

	input = &Input{
		Method: "total",
	}
	b, _ = json.Marshal(input)
	rc := new(chaincode.MockReadContext)
	rc.State = state
	rc.MockInput = b
	b, err = jrc.Query(rc)

	assert.NoError(err)

	var balance int64
	json.Unmarshal(b, &balance)

	assert.EqualValues(100, balance)

	input = &Input{
		Method: "balance",
		Dest:   []byte{2, 2, 2},
	}
	b, _ = json.Marshal(input)
	rc.MockInput = b
	b, err = jrc.Query(rc)

	assert.NoError(err)

	balance = 0
	json.Unmarshal(b, &balance)

	assert.EqualValues(100, balance)
}

func TestJuriaCoin_Transfer(t *testing.T) {
	assert := assert.New(t)
	state := chaincode.NewMockState()
	jrc := new(JuriaCoin)

	wc := new(chaincode.MockWriteContext)
	wc.MockSender = []byte{1, 1, 1}
	wc.State = state
	jrc.Init(wc)

	input := &Input{
		Method: "mint",
		Dest:   []byte{2, 2, 2},
		Value:  100,
	}
	b, _ := json.Marshal(input)
	wc.MockInput = b
	jrc.Invoke(wc)

	// transfer 222 -> 333, value = 101
	input = &Input{
		Method: "transfer",
		Dest:   []byte{3, 3, 3},
		Value:  101,
	}
	b, _ = json.Marshal(input)
	wc.MockSender = []byte{2, 2, 2}
	wc.MockInput = b
	err := jrc.Invoke(wc)

	assert.Error(err, "not enough coin error")

	// transfer 222 -> 333, value = 100
	input.Value = 100
	b, _ = json.Marshal(input)
	wc.MockInput = b
	err = jrc.Invoke(wc)

	assert.NoError(err)

	input.Method = "total"
	b, _ = json.Marshal(input)
	rc := new(chaincode.MockReadContext)
	rc.State = state
	rc.MockInput = b
	b, _ = jrc.Query(rc)
	var balance int64
	json.Unmarshal(b, &balance)

	assert.EqualValues(100, balance, "total should not change")

	input.Method = "balance"
	input.Dest = []byte{2, 2, 2}
	b, _ = json.Marshal(input)
	rc.MockInput = b
	b, _ = jrc.Query(rc)
	balance = 0
	json.Unmarshal(b, &balance)

	assert.EqualValues(0, balance)

	input.Dest = []byte{3, 3, 3}
	b, _ = json.Marshal(input)
	rc.MockInput = b
	b, _ = jrc.Query(rc)
	balance = 0
	json.Unmarshal(b, &balance)

	assert.EqualValues(100, balance)
}
