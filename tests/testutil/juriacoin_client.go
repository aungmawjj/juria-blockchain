// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package testutil

import (
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/execution"
	"github.com/aungmawjj/juria-blockchain/execution/chaincode/juriacoin"
	"github.com/aungmawjj/juria-blockchain/tests/cluster"
)

type JuriaCoinClient struct {
	minter   *core.PrivateKey
	accounts []*core.PrivateKey

	cluster  *cluster.Cluster
	codeAddr []byte

	transferIdx int64
}

var _ LoadClient = (*JuriaCoinClient)(nil)

// create and setup a LoadService
// submit chaincode deploy tx and wait for commit
func NewJuriaCoinLoadClient(mint int) *JuriaCoinClient {
	client := &JuriaCoinClient{
		minter:   core.GenerateKey(nil),
		accounts: make([]*core.PrivateKey, mint),
	}
	for i := 0; i < mint; i++ {
		client.accounts[i] = core.GenerateKey(nil)
	}
	return client
}

func (client *JuriaCoinClient) SetupOnCluster(cls *cluster.Cluster) error {
	return client.setupOnCluster(cls)
}

func (client *JuriaCoinClient) SubmitTxAndWait() error {
	return SubmitTxAndWait(client.cluster, client.makeRandomTransfer())
}

func (client *JuriaCoinClient) SubmitTx() (int, *core.Transaction, error) {
	tx := client.makeRandomTransfer()
	nodeIdx, err := SubmitTx(client.cluster, tx)
	return nodeIdx, tx, err
}

func (client *JuriaCoinClient) setupOnCluster(cls *cluster.Cluster) error {
	client.cluster = cls
	depTx := client.MakeDeploymentTx(client.minter)
	if err := SubmitTxAndWait(client.cluster, depTx); err != nil {
		return fmt.Errorf("cannot deploy juriacoin %w", err)
	}
	client.codeAddr = depTx.Hash()
	client.mintAccounts()
	return nil
}

func (client *JuriaCoinClient) mintAccounts() error {
	errCh := make(chan error, 10)
	for _, acc := range client.accounts {
		go func(acc *core.PublicKey) {
			errCh <- client.mintSingleAccount(acc)
		}(acc.PublicKey())
	}
	for range client.accounts {
		err := <-errCh
		if err != nil {
			return err
		}
	}
	return nil
}

func (client *JuriaCoinClient) mintSingleAccount(dest *core.PublicKey) error {
	var mintAmount int64 = 10000000000
	mintTx := client.MakeMintTx(dest, mintAmount)
	if err := SubmitTxAndWait(client.cluster, mintTx); err != nil {
		return fmt.Errorf("cannot mint juriacoin %w", err)
	}
	balance, err := client.QueryBalance(client.minter.PublicKey())
	if err != nil {
		return fmt.Errorf("cannot query juriacoin balance %w", err)
	}
	if mintAmount != balance {
		return fmt.Errorf("incorrect balance %d %d", mintAmount, balance)
	}
	return nil
}

func (client *JuriaCoinClient) makeRandomTransfer() *core.Transaction {
	i := int(atomic.AddInt64(&client.transferIdx, 1))
	if i >= len(client.accounts) {
		atomic.StoreInt64(&client.transferIdx, 0)
		i = 0
	}
	return client.MakeTransferTx(client.accounts[i],
		core.GenerateKey(nil).PublicKey(), 1)
}

func (client *JuriaCoinClient) QueryBalance(dest *core.PublicKey) (int64, error) {
	result, err := QueryState(client.cluster, client.MakeBalanceQuery(dest))
	if err != nil {
		return 0, err
	}
	var balance int64
	return balance, json.Unmarshal(result, &balance)
}

func (client *JuriaCoinClient) MakeDeploymentTx(minter *core.PrivateKey) *core.Transaction {
	input := &execution.DeploymentInput{
		CodeInfo: execution.CodeInfo{
			DriverType: execution.DriverTypeNative,
			CodeID:     execution.NativeCodeIDJuriaCoin,
		},
	}
	b, _ := json.Marshal(input)
	return core.NewTransaction().
		SetNonce(time.Now().UnixNano()).
		SetInput(b).
		Sign(minter)
}

func (client *JuriaCoinClient) MakeMintTx(dest *core.PublicKey, value int64) *core.Transaction {
	input := &juriacoin.Input{
		Method: "mint",
		Dest:   dest.Bytes(),
		Value:  value,
	}
	b, _ := json.Marshal(input)
	return core.NewTransaction().
		SetCodeAddr(client.codeAddr).
		SetNonce(time.Now().UnixNano()).
		SetInput(b).
		Sign(client.minter)
}

func (client *JuriaCoinClient) MakeTransferTx(
	sender *core.PrivateKey, dest *core.PublicKey, value int64,
) *core.Transaction {
	input := &juriacoin.Input{
		Method: "transfer",
		Dest:   dest.Bytes(),
		Value:  value,
	}
	b, _ := json.Marshal(input)
	return core.NewTransaction().
		SetCodeAddr(client.codeAddr).
		SetNonce(time.Now().UnixNano()).
		SetInput(b).
		Sign(sender)
}

func (client *JuriaCoinClient) MakeBalanceQuery(dest *core.PublicKey) *execution.QueryData {
	input := &juriacoin.Input{
		Method: "balance",
		Dest:   dest.Bytes(),
	}
	b, _ := json.Marshal(input)
	return &execution.QueryData{
		CodeAddr: client.codeAddr,
		Input:    b,
	}
}
