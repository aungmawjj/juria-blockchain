// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package testutil

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/execution"
	"github.com/aungmawjj/juria-blockchain/tests/cluster"
	"github.com/aungmawjj/juria-blockchain/txpool"
)

func SubmitTxAndWait(cls *cluster.Cluster, tx *core.Transaction) (int, error) {
	idx, err := SubmitTx(cls, tx)
	if err != nil {
		return 0, err
	}
	start := time.Now()
	for {
		status, err := GetTxStatus(cls.GetNode(idx), tx.Hash())
		if err != nil {
			return 0, fmt.Errorf("get tx status error %w", err)
		} else {
			if status == txpool.TxStatusNotFound {
				return 0, fmt.Errorf("submited tx status not found")
			}
			if status == txpool.TxStatusCommited {
				return idx, nil
			}
		}
		if time.Since(start) > 1*time.Second {
			// maybe current leader doesn't receive tx
			// resubmit tx again
			return SubmitTxAndWait(cls, tx)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func SubmitTx(cls *cluster.Cluster, tx *core.Transaction) (int, error) {
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
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
			return i, nil
		}
	}
	return 0, fmt.Errorf("cannot submit tx %w", retErr)
}

func GetTxStatus(node cluster.Node, hash []byte) (txpool.TxStatus, error) {
	hashstr := hex.EncodeToString(hash)
	resp, err := getRequestWithRetry(node.GetEndpoint() +
		fmt.Sprintf("/transactions/%s/status", hashstr))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var status txpool.TxStatus
	return status, json.NewDecoder(resp.Body).Decode(&status)
}

func QueryState(node cluster.Node, query *execution.QueryData) ([]byte, error) {
	b, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(node.GetEndpoint()+"/querystate",
		"application/json", bytes.NewReader(b))
	err = checkResponse(resp, err)
	if err != nil {
		return nil, fmt.Errorf("cannot query state %w", err)
	}
	defer resp.Body.Close()
	var ret []byte
	return ret, json.NewDecoder(resp.Body).Decode(&ret)
}

func uploadBinChainCode(cls *cluster.Cluster, binccPath string) (int, []byte, error) {
	buf, contentType, err := createBinccRequestBody(binccPath)
	if err != nil {
		return 0, nil, err
	}
	var retErr error
	retryOrder := PickUniqueRandoms(cls.NodeCount(), cls.NodeCount())
	for _, i := range retryOrder {
		if !cls.GetNode(i).IsRunning() {
			continue
		}
		resp, err := http.Post(cls.GetNode(i).GetEndpoint()+"/bincc",
			contentType, buf)
		retErr = checkResponse(resp, err)
		if retErr == nil {
			defer resp.Body.Close()
			var codeID []byte
			return i, codeID, json.NewDecoder(resp.Body).Decode(&codeID)
		}
	}
	return 0, nil, fmt.Errorf("cannot upload bincc %w", retErr)
}

func createBinccRequestBody(binccPath string) (*bytes.Buffer, string, error) {
	f, err := os.Open(binccPath)
	if err != nil {
		return nil, "", err
	}
	defer f.Close()

	buf := bytes.NewBuffer(nil)
	mw := multipart.NewWriter(buf)
	defer mw.Close()

	fw, err := mw.CreateFormFile("file", "binChaincode")
	if err != nil {
		return nil, "", err
	}
	if _, err := io.Copy(fw, f); err != nil {
		return nil, "", err
	}
	return buf, mw.FormDataContentType(), nil
}
