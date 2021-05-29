// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package testutil

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/execution"
	"github.com/aungmawjj/juria-blockchain/execution/chaincode/juriacoin"
	"github.com/aungmawjj/juria-blockchain/node"
	"github.com/aungmawjj/juria-blockchain/tests/cluster"
)

type JuriaCoinClient struct {
	minter   *core.PrivateKey
	accounts []*core.PrivateKey

	cluster  *cluster.Cluster
	codeAddr []byte
}

// create and setup a LoadService
// submit chaincode deploy tx and wait for commit
func NewJuriaCoinClient(mint int) *JuriaCoinClient {
	svc := &JuriaCoinClient{
		minter:   core.GenerateKey(nil),
		accounts: make([]*core.PrivateKey, mint),
	}
	for i := 0; i < mint; i++ {
		svc.accounts[i] = core.GenerateKey(nil)
	}
	return svc
}

func (svc *JuriaCoinClient) SetupOnCluster(cls *cluster.Cluster) error {
	return svc.setupOnCluster(cls)
}

func (svc *JuriaCoinClient) Transfer() error {
	return svc.transfer()
}

func (svc *JuriaCoinClient) TransferAsync() (*core.Transaction, error) {
	return svc.transferAsync()
}

func (svc *JuriaCoinClient) setupOnCluster(cls *cluster.Cluster) error {
	svc.cluster = cls
	depTx := svc.makeDeploymentTx()
	if err := submitTxAndWait(svc.cluster, depTx); err != nil {
		return fmt.Errorf("cannot deploy juriacoin %w", err)
	}
	svc.codeAddr = depTx.Hash()
	svc.mintAccounts()
	return nil
}

func (svc *JuriaCoinClient) mintAccounts() error {
	errCh := make(chan error, 10)
	for _, acc := range svc.accounts {
		go func(acc *core.PublicKey) {
			errCh <- svc.mintSingleAccount(acc)
		}(acc.PublicKey())
	}
	for range svc.accounts {
		err := <-errCh
		if err != nil {
			return err
		}
	}
	return nil
}

func (svc *JuriaCoinClient) mintSingleAccount(dest *core.PublicKey) error {
	var mintAmount int64 = 10000000000
	mintTx := svc.makeMintTx(dest, mintAmount)
	if err := submitTxAndWait(svc.cluster, mintTx); err != nil {
		return fmt.Errorf("cannot mint juriacoin %w", err)
	}
	balance, err := svc.queryBalance(svc.minter.PublicKey())
	if err != nil {
		return fmt.Errorf("cannot query juriacoin balance %w", err)
	}
	if mintAmount != balance {
		return fmt.Errorf("incorrect balance %d %d", mintAmount, balance)
	}
	return nil
}

func (svc *JuriaCoinClient) makeDeploymentTx() *core.Transaction {
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
		Sign(svc.minter)
}

func (svc *JuriaCoinClient) transfer() error {
	return submitTxAndWait(svc.cluster, svc.makeRandomTransfer())
}

func (svc *JuriaCoinClient) transferAsync() (*core.Transaction, error) {
	tx := svc.makeRandomTransfer()
	_, err := submitTx(svc.cluster, tx)
	return tx, err
}

func (svc *JuriaCoinClient) makeRandomTransfer() *core.Transaction {
	return svc.makeTransferTx(svc.accounts[rand.Intn(len(svc.accounts))],
		core.GenerateKey(nil).PublicKey(), 1)
}

func (svc *JuriaCoinClient) makeMintTx(dest *core.PublicKey, value int64) *core.Transaction {
	input := &juriacoin.Input{
		Method: "mint",
		Dest:   dest.Bytes(),
		Value:  value,
	}
	b, _ := json.Marshal(input)
	return core.NewTransaction().
		SetCodeAddr(svc.codeAddr).
		SetNonce(time.Now().UnixNano()).
		SetInput(b).
		Sign(svc.minter)
}

func (svc *JuriaCoinClient) makeTransferTx(
	sender *core.PrivateKey, dest *core.PublicKey, value int64,
) *core.Transaction {
	input := &juriacoin.Input{
		Method: "transfer",
		Dest:   dest.Bytes(),
		Value:  value,
	}
	b, _ := json.Marshal(input)
	return core.NewTransaction().
		SetCodeAddr(svc.codeAddr).
		SetNonce(time.Now().UnixNano()).
		SetInput(b).
		Sign(sender)
}

func (svc *JuriaCoinClient) queryBalance(dest *core.PublicKey) (int64, error) {
	input := &juriacoin.Input{
		Method: "balance",
		Dest:   dest.Bytes(),
	}
	b, _ := json.Marshal(input)
	query := &node.StateQuery{
		CodeAddr: svc.codeAddr,
		Input:    b,
	}
	result, err := queryState(svc.cluster, query)
	if err != nil {
		return 0, err
	}
	var balance int64
	return balance, json.Unmarshal(result, &balance)
}
