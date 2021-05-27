// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package testutil

import (
	"encoding/json"
	"fmt"
	"net/http"

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

func GetConsensusStatus(node cluster.Node) (*consensus.Status, error) {
	resp, err := http.Get(node.GetEndpoint() + "/consensus")
	if err := checkResponse(resp, err); err != nil {
		return nil, err
	}
	ret := new(consensus.Status)
	if err := json.NewDecoder(resp.Body).Decode(ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func GetBlockByHeight(node cluster.Node, height uint64) (*core.Block, error) {
	resp, err := http.Get(fmt.Sprintf("%s/blocksbyh/%d", node.GetEndpoint(), height))
	if err := checkResponse(resp, err); err != nil {
		return nil, err
	}
	ret := core.NewBlock()
	if err := json.NewDecoder(resp.Body).Decode(ret); err != nil {
		return nil, err
	}
	return ret, nil
}
