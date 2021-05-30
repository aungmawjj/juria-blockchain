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
)

func HealthCheckAll(cls *cluster.Cluster) error {
	fmt.Println("Health check all nodes")
	hc := &healthChecker{
		cluster:  cls,
		majority: false,
	}
	return hc.run()
}

func HealthCheckMajority(cls *cluster.Cluster) error {
	fmt.Println("Health check majority nodes")
	hc := &healthChecker{
		cluster:  cls,
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
get status, remember heighest bexec and commitedTxCount
wait for 20s
for majority check, wait more for ((total - majority) * leaderTimeout)
get status again
bexec must be higher than previous one
should get txCommit with txHash from nodes

Rotation
make a timeout channel for (viewWidth + 5s)
for majority check, add ((total - majority) * leaderTimeout) to timeout duration
get status every 1s
on each node leader change must occur before timeout
after leader change, all leaderIdx should be equal
*/
type healthChecker struct {
	cluster  *cluster.Cluster
	majority bool // should (majority or all) nodes healthy

	interrupt chan struct{}
	mtxIntr   sync.Mutex
	err       error
}

func (hc *healthChecker) run() error {
	hc.interrupt = make(chan struct{})

	wg := new(sync.WaitGroup)
	wg.Add(3)
	go hc.runChecker(hc.checkSafety, wg)
	go hc.runChecker(hc.checkLiveness, wg)
	go hc.runChecker(hc.checkRotation, wg)
	wg.Wait()
	return hc.err
}

func (hc *healthChecker) runChecker(checker func() error, wg *sync.WaitGroup) {
	defer wg.Done()
	err := checker()
	if err != nil {
		hc.err = err
		hc.makeInterrupt()
	}
}

func (hc *healthChecker) makeInterrupt() {
	hc.mtxIntr.Lock()
	defer hc.mtxIntr.Unlock()

	select {
	case <-hc.interrupt:
		return
	default:
	}
	close(hc.interrupt)
}

func (hc *healthChecker) checkSafety() error {
	height, err := hc.getMinimumBexec()
	if err != nil {
		return err
	}
	select {
	case <-hc.interrupt:
		return nil
	default:
	}
	return hc.shouldEqualMerkleRoot(height)
}

func (hc *healthChecker) shouldEqualMerkleRoot(height uint64) error {
	bResp, err := hc.shouldGetBlockByHeight(height)
	if err != nil {
		return err
	}
	var mRoot []byte
	equalCount := 1
	for _, blk := range bResp {
		if blk.MerkleRoot() == nil {
			equalCount++
		} else if mRoot == nil {
			mRoot = blk.MerkleRoot()
		} else if bytes.Equal(mRoot, blk.MerkleRoot()) {
			equalCount++
		}
	}
	if equalCount < hc.minimumHealthyNode() {
		return fmt.Errorf("different merkle root at block %d", height)
	}
	return nil
}

func (hc *healthChecker) getMinimumBexec() (uint64, error) {
	sResp, err := hc.shouldGetStatus()
	if err != nil {
		return 0, err
	}
	var ret uint64 = 0
	for _, status := range sResp {
		if ret == 0 || status.BExec < ret {
			ret = status.BExec
		}
	}
	return ret, nil
}

func (hc *healthChecker) checkLiveness() error {
	status, err := hc.shouldGetStatus()
	if err != nil {
		return err
	}
	lastHeight := hc.getMaximumBexec(status)
	time.Sleep(hc.getLivenessWaitTime())
	select {
	case <-hc.interrupt:
		return nil
	default:
	}
	prevStatus := status
	status, err = hc.shouldGetStatus()
	if err != nil {
		return err
	}
	if err := hc.shouldCommitNewBlocks(status, lastHeight); err != nil {
		return err
	}
	return hc.shouldCommitTxs(prevStatus, status)
}

func (hc *healthChecker) getMaximumBexec(status map[int]*consensus.Status) uint64 {
	var bexec uint64 = 0
	var txCount int = 0
	for _, s := range status {
		if s.BExec > bexec {
			bexec = s.BExec
		}
		if s.CommitedTxCount > txCount {
			txCount = s.CommitedTxCount
		}
	}
	return bexec
}

func (hc *healthChecker) getLivenessWaitTime() time.Duration {
	d := 20 * time.Second
	if hc.majority {
		d += time.Duration(hc.getFaultyCount()) * LeaderTimeout()
	}
	return d
}

func (hc *healthChecker) shouldCommitNewBlocks(
	prevStatus map[int]*consensus.Status, lastHeight uint64,
) error {
	validCount := 0
	for _, status := range prevStatus {
		if status.BExec > lastHeight {
			validCount++
		}
	}
	if validCount < hc.minimumHealthyNode() {
		return fmt.Errorf("%d nodes are not commiting new blocks",
			hc.cluster.NodeCount()-validCount)
	}
	return nil
}

func (hc *healthChecker) shouldCommitTxs(
	prevStatus, status map[int]*consensus.Status,
) error {
	validCount := 0
	for i, s := range status {
		if prevStatus == nil && s.CommitedTxCount > 0 {
			validCount++
		} else if s.CommitedTxCount > prevStatus[i].CommitedTxCount {
			validCount++
		}
	}
	if validCount < hc.minimumHealthyNode() {
		return fmt.Errorf("%d nodes are not commiting new txs",
			hc.cluster.NodeCount()-validCount)
	}
	return nil
}

func (hc *healthChecker) getFaultyCount() int {
	return hc.cluster.NodeCount() - core.MajorityCount(hc.cluster.NodeCount())
}

func (hc *healthChecker) checkRotation() error {
	timeout := time.NewTimer(hc.getRotationTimeout())
	defer timeout.Stop()

	lastView := make(map[int]*consensus.Status)
	changedView := make(map[int]*consensus.Status)
	for {
		if err := hc.updateViewChangeStatus(lastView, changedView); err != nil {
			return err
		}
		if len(changedView) >= len(lastView) {
			return hc.shouldEqualLeader(changedView)
		}
		select {
		case <-hc.interrupt:
			return nil

		case <-timeout.C:
			return fmt.Errorf("cluster failed to rotate leader")

		case <-time.After(1 * time.Second):
		}
	}
}

func (hc *healthChecker) getRotationTimeout() time.Duration {
	d := consensus.DefaultConfig.ViewWidth + 5*time.Second
	if hc.majority {
		d += time.Duration(hc.getFaultyCount()) * LeaderTimeout()
	}
	return d
}

func (hc *healthChecker) updateViewChangeStatus(last, changed map[int]*consensus.Status) error {
	sResp, err := hc.shouldGetStatus()
	if err != nil {
		return err
	}
	for i, status := range sResp {
		if hc.hasViewChanged(status, last[i]) {
			changed[i] = status
			last[i] = status
		}
		if last[i] == nil {
			last[i] = status // first time
		}
	}
	return nil
}

func (hc *healthChecker) hasViewChanged(status, last *consensus.Status) bool {
	if last == nil {
		return false // last view is not loaded yet for the first time
	}
	if status.PendingViewChange {
		return false // view change is pending but not confirmed yet
	}
	if last.PendingViewChange { // previously pending, now not pending
		return true // it's confirmed view change
	}
	return status.LeaderIndex != last.LeaderIndex
}

func (hc *healthChecker) shouldEqualLeader(changedView map[int]*consensus.Status) error {
	leaderIdx := -1
	equalCount := 1
	for _, status := range changedView {
		if leaderIdx == -1 {
			leaderIdx = status.LeaderIndex
		} else if leaderIdx == status.LeaderIndex {
			equalCount++
		}
	}
	if equalCount < hc.minimumHealthyNode() {
		return fmt.Errorf("inconsistant view change")
	}
	return nil
}

func (hc *healthChecker) shouldGetStatus() (map[int]*consensus.Status, error) {
	ret := GetStatusAll(hc.cluster)
	min := hc.minimumHealthyNode()
	if len(ret) < min {
		return nil, fmt.Errorf("failed to get status from %d nodes", min-len(ret))
	}
	return ret, nil
}

func (hc *healthChecker) shouldGetBlockByHeight(height uint64) (map[int]*core.Block, error) {
	ret := GetBlockByHeightAll(hc.cluster, height)
	min := hc.minimumHealthyNode()
	if len(ret) < min {
		return nil, fmt.Errorf("failed to get block %d from %d nodes",
			height, min-len(ret))
	}
	return ret, nil
}

func (hc *healthChecker) minimumHealthyNode() int {
	min := hc.cluster.NodeCount()
	if hc.majority {
		min = core.MajorityCount(hc.cluster.NodeCount())
	}
	return min
}