// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package testutil

import (
	"encoding/json"
	"fmt"
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
		return fmt.Errorf("status code not 200")
	}
	return nil
}

func getRequestWithRetry(url string) (*http.Response, error) {
	retry := 0
	resp, err := http.Get(url)
	for {
		err = checkResponse(resp, err)
		if err == nil {
			return resp, err
		}
		retry++
		if retry > 3 {
			return nil, err
		}
		time.Sleep(5 * time.Millisecond)
		resp, err = http.Get(url)
	}
}

func GetStatus(node cluster.Node) (*consensus.Status, error) {
	resp, err := getRequestWithRetry(node.GetEndpoint() + "/consensus")
	if err != nil {
		return nil, err
	}
	ret := new(consensus.Status)
	if err := json.NewDecoder(resp.Body).Decode(ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func GetStatusMany(cls *cluster.Cluster, min int) (map[int]*consensus.Status, error) {
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
	if len(resps) < min {
		return nil, fmt.Errorf("cannot get status from %d nodes", min)
	}
	return resps, nil
}

func GetBlockByHeight(node cluster.Node, height uint64) (*core.Block, error) {
	resp, err := getRequestWithRetry(fmt.Sprintf("%s/blocksbyh/%d", node.GetEndpoint(), height))
	if err != nil {
		return nil, err
	}
	ret := core.NewBlock()
	if err := json.NewDecoder(resp.Body).Decode(ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func GetBlockByHeightMany(cls *cluster.Cluster, min int, height uint64) (map[int]*core.Block, error) {
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
	if len(resps) < min {
		return nil, fmt.Errorf("cannot get block by height from %d nodes", min)
	}
	return resps, nil
}
