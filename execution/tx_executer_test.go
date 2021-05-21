// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package execution

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/execution/chaincode/juriacoin"
	"github.com/stretchr/testify/assert"
)

func TestTxExecuter(t *testing.T) {
	assert := assert.New(t)

	priv := core.GenerateKey(nil)
	depInput := &DeploymentInput{
		CodeInfo: CodeInfo{
			DriverType: DriverTypeNative,
			CodeID:     []byte(NativeCodeIDJuriaCoin),
		},
	}
	b, _ := json.Marshal(depInput)
	txDep := core.NewTransaction().
		SetCodeAddr(nil).
		SetInput(b).
		Sign(priv)

	blk := core.NewBlock().SetHeight(10).Sign(priv)

	trk := newStateTracker(newMapStateStore(), nil)
	reg := newCodeRegistry()
	texe := txExecutor{
		codeRegistry: reg,
		timeout:      1 * time.Second,
		rootTrk:      trk,
		blk:          blk,
		tx:           txDep,
	}
	txc := texe.execute()

	assert.NotEqual("", txc.Error(), "code driver not registered")

	reg.registerDriver(DriverTypeNative, newNativeCodeDriver())
	txc = texe.execute()

	assert.Equal("", txc.Error())

	// codeinfo must be saved by key (transaction hash)
	cinfo, err := reg.getCodeInfo(txDep.Hash(), trk.spawn(codeRegistryAddr))

	assert.NoError(err)
	assert.Equal(*cinfo, depInput.CodeInfo)

	cc, err := reg.getInstance(txDep.Hash(), trk.spawn(codeRegistryAddr))

	assert.NoError(err)
	assert.NotNil(cc)

	ccInput := &juriacoin.Input{
		Method: "minter",
	}
	b, _ = json.Marshal(ccInput)
	minter, err := cc.Query(&callContext{
		input: b,
		State: trk.spawn(txDep.Hash()),
	})

	assert.NoError(err)
	assert.Equal(priv.PublicKey().Bytes(), minter, "deployer must be set as minter")

	ccInput.Method = "mint"
	ccInput.Dest = priv.PublicKey().Bytes()
	ccInput.Value = 100
	b, _ = json.Marshal(ccInput)

	txInvoke := core.NewTransaction().
		SetCodeAddr(txDep.Hash()).
		SetInput(b).
		Sign(priv)

	texe.tx = txInvoke
	txc = texe.execute()

	assert.Equal("", txc.Error())

	ccInput.Method = "balance"
	ccInput.Value = 0
	b, _ = json.Marshal(ccInput)

	b, err = cc.Query(&callContext{
		input: b,
		State: trk.spawn(txDep.Hash()),
	})

	var balance int64
	json.Unmarshal(b, &balance)

	assert.NoError(err)
	assert.EqualValues(100, balance)
}
