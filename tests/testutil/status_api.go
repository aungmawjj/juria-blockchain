// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package testutil

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/aungmawjj/juria-blockchain/consensus"
	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/tests/cluster"
)

func checkResponse(resp *http.Response, err error) error {
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		msg, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		return fmt.Errorf("status code not 200 %s", string(msg))
	}
	return nil
}

func getRequestWithRetry(url string) (*http.Response, error) {
	retry := 0
	for {
		resp, err := http.Get(url)
		err = checkResponse(resp, err)
		if err == nil {
			return resp, err
		}
		fmt.Println("get request error", err)
		retry++
		if retry > 5 {
			return nil, err
		}
		time.Sleep(1 * time.Second)
	}
}

func GetStatus(node cluster.Node) (*consensus.Status, error) {
	if !node.IsRunning() {
		return nil, fmt.Errorf("node is not running")
	}
	resp, err := getRequestWithRetry(node.GetEndpoint() + "/consensus")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	ret := new(consensus.Status)
	if err := json.NewDecoder(resp.Body).Decode(ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func GetStatusAll(cls *cluster.Cluster) map[int]*consensus.Status {
	resps := make(map[int]*consensus.Status)
	var mtx sync.Mutex
	var wg sync.WaitGroup
	wg.Add(cls.NodeCount())
	for i := 0; i < cls.NodeCount(); i++ {
		go func(i int) {
			defer wg.Done()
			resp, err := GetStatus(cls.GetNode(i))
			if err == nil {
				mtx.Lock()
				defer mtx.Unlock()
				resps[i] = resp
			}
		}(i)
	}
	wg.Wait()
	return resps
}

func GetBlockByHeight(node cluster.Node, height uint64) (*core.Block, error) {
	if !node.IsRunning() {
		return nil, fmt.Errorf("node is not running")
	}
	resp, err := getRequestWithRetry(fmt.Sprintf("%s/blocksbyh/%d", node.GetEndpoint(), height))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	ret := core.NewBlock()
	if err := json.NewDecoder(resp.Body).Decode(ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func GetBlockByHeightAll(cls *cluster.Cluster, height uint64) map[int]*core.Block {
	resps := make(map[int]*core.Block)
	var mtx sync.Mutex
	var wg sync.WaitGroup
	wg.Add(cls.NodeCount())
	for i := 0; i < cls.NodeCount(); i++ {
		go func(i int) {
			defer wg.Done()
			resp, err := GetBlockByHeight(cls.GetNode(i), height)
			if err == nil {
				mtx.Lock()
				defer mtx.Unlock()
				resps[i] = resp
			}
		}(i)
	}
	wg.Wait()
	return resps
}
