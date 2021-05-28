// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package testutil

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/execution"
	"github.com/aungmawjj/juria-blockchain/tests/cluster"
)

type LoadJuriaCoin struct {
	cluster *cluster.Cluster

	minter   *core.PrivateKey
	accounts []*core.PrivateKey
}

// create and setup a LoadService
// submit chaincode deploy tx and wait for commit
func NewLoadService(cls *cluster.Cluster) (*LoadJuriaCoin, error) {
	svc := &LoadJuriaCoin{
		cluster: cls,
	}
	err := svc.setup()
	if err != nil {
		return nil, err
	}
	return svc, nil
}

// submit tx and wait for commit
func (svc *LoadJuriaCoin) SubmitTxAndWait(ctx context.Context) error {
	return nil
}

func (svc *LoadJuriaCoin) setup() error {
	svc.minter = core.GenerateKey(nil)
	return nil
}

func (svc *LoadJuriaCoin) makeDeploymentTx() *core.Transaction {
	input := execution.DeploymentInput{
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

func (svc *LoadJuriaCoin) makeMintTx() (*core.Transaction, error) {
	return nil, nil
}

func (svc *LoadJuriaCoin) makeTransferTx() (*core.Transaction, error) {
	return nil, nil
}
