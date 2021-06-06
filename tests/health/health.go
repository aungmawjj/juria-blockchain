// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package health

import (
	"fmt"
	"sync"
	"time"

	"github.com/aungmawjj/juria-blockchain/consensus"
	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/tests/cluster"
	"github.com/aungmawjj/juria-blockchain/tests/testutil"
)

func CheckAllNodes(cls *cluster.Cluster) error {
	fmt.Println("Health check all nodes")
	hc := &checker{
		cluster:  cls,
		majority: false,
	}
	return hc.run()
}

func CheckMajorityNodes(cls *cluster.Cluster) error {
	fmt.Println("Health check majority nodes")
	hc := &checker{
		cluster:  cls,
		majority: true,
	}
	return hc.run()
}

/*
checker check cluster's health in three aspects

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
type checker struct {
	cluster  *cluster.Cluster
	majority bool // should (majority or all) nodes healthy

	interrupt chan struct{}
	mtxIntr   sync.Mutex
	err       error
}

func (hc *checker) run() error {
	hc.interrupt = make(chan struct{})

	wg := new(sync.WaitGroup)
	wg.Add(3)
	go hc.runChecker(hc.checkSafety, wg)
	go hc.runChecker(hc.checkLiveness, wg)
	go hc.runChecker(hc.checkRotation, wg)
	wg.Wait()
	return hc.err
}

func (hc *checker) runChecker(checker func() error, wg *sync.WaitGroup) {
	defer wg.Done()
	err := checker()
	if err != nil {
		hc.err = err
		hc.makeInterrupt()
	}
}

func (hc *checker) makeInterrupt() {
	hc.mtxIntr.Lock()
	defer hc.mtxIntr.Unlock()

	select {
	case <-hc.interrupt:
		return
	default:
	}
	close(hc.interrupt)
}

func (hc *checker) shouldGetStatus() (map[int]*consensus.Status, error) {
	ret := testutil.GetStatusAll(hc.cluster)
	min := hc.minimumHealthyNode()
	if len(ret) < min {
		return nil, fmt.Errorf("failed to get status from %d nodes", min-len(ret))
	}
	return ret, nil
}

func (hc *checker) minimumHealthyNode() int {
	min := hc.cluster.NodeCount()
	if hc.majority {
		min = core.MajorityCount(hc.cluster.NodeCount())
	}
	return min
}

func (hc *checker) LeaderTimeout() time.Duration {
	config := hc.cluster.NodeConfig()
	return config.ConsensusConfig.LeaderTimeout
}

func (hc *checker) getFaultyCount() int {
	return hc.cluster.NodeCount() - core.MajorityCount(hc.cluster.NodeCount())
}
