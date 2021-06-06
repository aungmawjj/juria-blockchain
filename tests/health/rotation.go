// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package health

import (
	"fmt"
	"time"

	"github.com/aungmawjj/juria-blockchain/consensus"
)

func (hc *checker) checkRotation() error {
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

func (hc *checker) getRotationTimeout() time.Duration {
	config := hc.cluster.NodeConfig()
	d := config.ConsensusConfig.ViewWidth + 5*time.Second
	if hc.majority {
		d += time.Duration(hc.getFaultyCount()) * hc.LeaderTimeout()
	}
	return d
}

func (hc *checker) updateViewChangeStatus(last, changed map[int]*consensus.Status) error {
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

func (hc *checker) hasViewChanged(status, last *consensus.Status) bool {
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

func (hc *checker) shouldEqualLeader(changedView map[int]*consensus.Status) error {
	equalCount := make(map[int]int)
	for _, status := range changedView {
		equalCount[status.LeaderIndex]++
	}
	for i, count := range equalCount {
		if count >= hc.minimumHealthyNode() {
			fmt.Printf(" + Leader changed to %d\n", i)
			return nil
		}
	}
	return fmt.Errorf("inconsistant view change")
}
