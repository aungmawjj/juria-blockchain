// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package testutil

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/aungmawjj/juria-blockchain/consensus"
	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/tests/cluster"
	"github.com/fatih/color"
)

func HealthCheckAll(cls *cluster.Cluster) error {
	fmt.Println("Health check all nodes")
	hc := &healthChecker{
		cls:      cls,
		majority: false,
	}
	return hc.run()
}

func HealthCheckMajority(cls *cluster.Cluster) error {
	fmt.Println("Health check majority nodes")
	hc := &healthChecker{
		cls:      cls,
		majority: true,
	}
	return hc.run()
}

/*
healthChecker check cluster's health in three aspects

Safety
get status, select lowest bexec height
get bexec block, all bexec.MerkleRoot must be equal

Liveness
get status, remember heighest bexec
wait for leaderTimeout duration
for majority check, wait more for ((total - majority) * leaderTimeout)
get status again
bexec must be higher than previous one
should get txCommit with txHash from nodes

Rotation
make a timeout channel for (viewWidth + leaderTimeout)
get status every 3s
on each node leader change must occur before timeout
after leader change, all leaderIdx should be equal
*/
type healthChecker struct {
	cls       *cluster.Cluster
	majority  bool // should (majority or all) nodes healthy
	interrupt chan struct{}
}

func (hc *healthChecker) run() error {
	err := hc.runParallel()
	if err != nil {
		color.Red("Health check FAIL: %s", err)
	} else {
		color.Green("Health check PASS")
	}
	return err
}

func (hc *healthChecker) runParallel() error {
	hc.interrupt = make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(3)
	var err error
	go func() {
		defer wg.Done()
		err = hc.checkSafety()
		if err != nil {
			hc.makeInterrupt()
		}
	}()
	go func() {
		defer wg.Done()
		err = hc.checkLiveness()
		if err != nil {
			hc.makeInterrupt()
		}
	}()
	go func() {
		defer wg.Done()
		err = hc.checkRotation()
		if err != nil {
			hc.makeInterrupt()
		}
	}()
	wg.Wait()
	return err
}

func (hc *healthChecker) makeInterrupt() {
	select {
	case <-hc.interrupt:
		return
	default:
	}
	close(hc.interrupt)
}

func (hc *healthChecker) checkSafety() error {
	sResp, err := hc.shouldGetStatus()
	if err != nil {
		return err
	}
	var height uint64 = 0 // lowest committed block height on all nodes
	for _, status := range sResp {
		if height == 0 || status.BExec < height {
			height = status.BExec
		}
	}
	select {
	case <-hc.interrupt:
		return nil
	default:
	}
	bResp, err := hc.shouldGetBlockByHeight(height)
	if err != nil {
		return err
	}
	var mRoot []byte
	for _, blk := range bResp {
		if mRoot == nil {
			mRoot = blk.MerkleRoot()
			continue
		}
		if !bytes.Equal(mRoot, blk.MerkleRoot()) {
			return fmt.Errorf("different merkle root at block %d", height)
		}
	}
	return nil
}

func (hc *healthChecker) checkLiveness() error {
	sResp, err := hc.shouldGetStatus()
	if err != nil {
		return err
	}
	var lastHeight uint64 = 0 // highest committed block height on all nodes
	for _, status := range sResp {
		if status.BExec > lastHeight {
			lastHeight = status.BExec
		}
	}
	time.Sleep(consensus.DefaultConfig.LeaderTimeout)
	if hc.majority {
		maxFaulty := hc.cls.NodeCount() - core.MajorityCount(hc.cls.NodeCount())
		time.Sleep(time.Duration(maxFaulty) * consensus.DefaultConfig.LeaderTimeout)
	}
	select {
	case <-hc.interrupt:
		return nil
	default:
	}
	sResp2, err := hc.shouldGetStatus()
	if err != nil {
		return err
	}
	for i, status := range sResp2 {
		if status.BExec <= lastHeight {
			return fmt.Errorf("node %d is not commiting new blocks", i)
		}
	}
	return nil
}

func (hc *healthChecker) checkRotation() error {
	timeout := time.NewTimer(consensus.DefaultConfig.ViewWidth + 10*time.Second)
	defer timeout.Stop()

	lastView := make(map[int]*consensus.Status)
	changedView := make(map[int]*consensus.Status)

	for {
		select {
		case <-hc.interrupt:
			return nil

		case <-timeout.C:
			return fmt.Errorf("cluster failed to rotate leader")

		case <-time.After(3 * time.Second):
		}

		sResp, err := hc.shouldGetStatus()
		if err != nil {
			return err
		}
		for i, status := range sResp {
			if hc.hasViewChanged(status, lastView[i]) {
				changedView[i] = status
				lastView[i] = status
			}
			if lastView[i] == nil {
				lastView[i] = status
			}
		}
		if len(changedView) >= hc.minimumHealthyNode() {
			leaderIdx := -1
			for _, status := range changedView {
				if leaderIdx == -1 {
					leaderIdx = status.LeaderIndex
				} else if leaderIdx != status.LeaderIndex {
					return fmt.Errorf("inconsistant view change")
				}
			}
			return nil
		}

	}
}

func (hc *healthChecker) hasViewChanged(status, prev *consensus.Status) bool {
	if prev == nil {
		return false
	}
	if status.PendingViewChange {
		return false
	}
	if prev.PendingViewChange {
		return true
	}
	return status.LeaderIndex != prev.LeaderIndex
}

func (hc *healthChecker) shouldGetStatus() (map[int]*consensus.Status, error) {
	return GetStatusMany(hc.cls, hc.minimumHealthyNode())
}

func (hc *healthChecker) shouldGetBlockByHeight(height uint64) (map[int]*core.Block, error) {
	return GetBlockByHeightMany(hc.cls, hc.minimumHealthyNode(), height)
}

func (hc *healthChecker) minimumHealthyNode() int {
	min := hc.cls.NodeCount()
	if hc.majority {
		min = core.MajorityCount(hc.cls.NodeCount())
	}
	return min
}
