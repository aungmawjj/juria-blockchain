// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package execution

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/stretchr/testify/assert"
)

func TestExecution(t *testing.T) {
	assert := assert.New(t)

	state := newMapStateStore()
	reg := newCodeRegistry()
	reg.registerDriver(DriverTypeNative, newNativeCodeDriver())

	execution := &Execution{
		stateStore:   state,
		codeRegistry: reg,
		config:       DefaultConfig,
	}
	execution.config.TxExecTimeout = 1 * time.Second

	priv := core.GenerateKey(nil)
	blk := core.NewBlock().SetHeight(10).Sign(priv)

	cinfo := CodeInfo{
		DriverType: DriverTypeNative,
		CodeID:     []byte(NativeCodeIDJuriaCoin),
	}
	cinfo2 := CodeInfo{
		DriverType: DriverTypeNative,
		CodeID:     []byte{2, 2, 2}, // invalid code id
	}

	depInput := &DeploymentInput{CodeInfo: cinfo}
	b, _ := json.Marshal(depInput)

	depInput.CodeInfo = cinfo2
	b2, _ := json.Marshal(depInput)

	tx1 := core.NewTransaction().SetNonce(time.Now().Unix()).SetInput(b).Sign(priv)
	tx2 := core.NewTransaction().SetNonce(time.Now().Unix()).SetInput(b2).Sign(priv)
	tx3 := core.NewTransaction().SetNonce(time.Now().Unix()).SetInput(b).Sign(priv)

	bcm, txcs := execution.Execute(blk, []*core.Transaction{tx1, tx2, tx3})

	assert.Equal(blk.Hash(), bcm.Hash())
	assert.EqualValues(3, len(txcs))
	assert.NotEmpty(bcm.StateChanges())

	// assert.Equal(tx1.Hash(), txcs[0].Hash())
	// assert.Equal(tx2.Hash(), txcs[1].Hash())
	// assert.Equal(tx3.Hash(), txcs[2].Hash())

	// for _, sc := range bcm.StateChanges() {
	// 	state.SetState(sc.Key(), sc.Value())
	// }

	// regTrk := newStateTracker(state, codeRegistryAddr)
	// resci, err := reg.getCodeInfo(tx1.Hash(), regTrk)

	// assert.NoError(err)
	// assert.Equal(&cinfo, resci)

	// resci, err = reg.getCodeInfo(tx2.Hash(), regTrk)

	// assert.Error(err)
	// assert.Nil(resci)

	// resci, err = reg.getCodeInfo(tx3.Hash(), regTrk)

	// assert.NoError(err)
	// assert.Equal(&cinfo, resci)

	// ccInput, _ := json.Marshal(juriacoin.Input{Method: "minter"})
	// minter, err := execution.Query(&QueryData{tx1.Hash(), ccInput})

	// assert.NoError(err)
	// assert.Equal(priv.PublicKey().Bytes(), minter)

	// minter, err = execution.Query(&QueryData{tx2.Hash(), ccInput})

	// assert.Error(err)
	// assert.Nil(minter)

	// minter, err = execution.Query(&QueryData{tx3.Hash(), ccInput})

	// assert.NoError(err)
	// assert.Equal(priv.PublicKey().Bytes(), minter)
}
