// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package health

import (
	"fmt"

	"github.com/aungmawjj/juria-blockchain/consensus"
	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/tests/testutil"
)

func (hc *checker) checkSafety() error {
	status, err := hc.shouldGetStatus()
	if err != nil {
		return err
	}
	height, err := hc.getMinimumBexec(status)
	if err != nil {
		return err
	}
	select {
	case <-hc.interrupt:
		return nil
	default:
	}
	blocks, err := hc.shouldGetBlockByHeight(height)
	if err != nil {
		return err
	}
	return hc.shouldEqualMerkleRoot(blocks)
}

func (hc *checker) getMinimumBexec(sMap map[int]*consensus.Status) (uint64, error) {
	var ret uint64
	for _, status := range sMap {
		if ret == 0 {
			ret = status.BExec
		} else if status.BExec < ret {
			ret = status.BExec
		}
	}
	return ret, nil
}

func (hc *checker) shouldGetBlockByHeight(height uint64) (map[int]*core.Block, error) {
	ret := testutil.GetBlockByHeightAll(hc.cluster, height)
	min := hc.minimumHealthyNode()
	if len(ret) < min {
		return nil, fmt.Errorf("failed to get block %d from %d nodes",
			height, min-len(ret))
	}
	return ret, nil
}

func (hc *checker) shouldEqualMerkleRoot(blocks map[int]*core.Block) error {
	var height uint64
	equalCount := make(map[string]int)
	for i, blk := range blocks {
		if blk.MerkleRoot() == nil {
			return fmt.Errorf("nil merkle root at node %d, block %d", i, blk.Height())
		}
		equalCount[string(blk.MerkleRoot())]++
		if height == 0 {
			height = blk.Height()
		}
	}
	for _, count := range equalCount {
		if count >= hc.minimumHealthyNode() {
			fmt.Printf(" + Same merkle root at block %d\n", height)
			return nil
		}
	}
	return fmt.Errorf("different merkle root at block %d", height)
}
