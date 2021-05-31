// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package testutil

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/execution"
	"github.com/aungmawjj/juria-blockchain/tests/cluster"
	"github.com/aungmawjj/juria-blockchain/txpool"
)

func submitTxAndWait(cls *cluster.Cluster, tx *core.Transaction) error {
	idx, err := submitTx(cls, tx)
	if err != nil {
		return err
	}
	errCount := 0
	for {
		status, err := getTxStatus(cls.GetNode(idx), tx.Hash())
		if err != nil {
			errCount++
			if errCount > 5 {
				return err
			}
			continue
		} else {
			if status == txpool.TxStatusNotFound {
				return fmt.Errorf("submited tx status not found")
			}
			if status == txpool.TxStatusCommited {
				return nil
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func submitTx(cls *cluster.Cluster, tx *core.Transaction) (int, error) {
	b, err := json.Marshal(tx)
	if err != nil {
		return 0, err
	}
	var retErr error
	retryOrder := PickUniqueRandoms(cls.NodeCount(), cls.NodeCount())
	for _, i := range retryOrder {
		if !cls.GetNode(i).IsRunning() {
			continue
		}
		resp, err := http.Post(cls.GetNode(i).GetEndpoint()+"/transactions",
			"application/json", bytes.NewReader(b))
		retErr = checkResponse(resp, err)
		if retErr == nil {
			return i, nil
		}
	}
	return 0, fmt.Errorf("cannot submit tx %w", retErr)
}

func getTxStatus(node cluster.Node, hash []byte) (txpool.TxStatus, error) {
	hashstr := hex.EncodeToString(hash)
	resp, err := getRequestWithRetry(node.GetEndpoint() +
		fmt.Sprintf("/transactions/%s/status", hashstr))
	if err != nil {
		return 0, err
	}

	var status txpool.TxStatus
	return status, json.NewDecoder(resp.Body).Decode(&status)
}

func queryState(cls *cluster.Cluster, query *execution.QueryData) ([]byte, error) {
	b, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}
	var retErr error
	retryOrder := PickUniqueRandoms(cls.NodeCount(), cls.NodeCount())
	for _, i := range retryOrder {
		if !cls.GetNode(i).IsRunning() {
			continue
		}
		resp, err := http.Post(cls.GetNode(i).GetEndpoint()+"/querystate",
			"application/json", bytes.NewReader(b))
		retErr = checkResponse(resp, err)
		if retErr == nil {
			var b []byte
			return b, json.NewDecoder(resp.Body).Decode(&b)
		}
	}
	return nil, fmt.Errorf("cannot query state %w", retErr)
}
