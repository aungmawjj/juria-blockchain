// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package experiments

import (
	"fmt"
	"time"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/tests/cluster"
	"github.com/aungmawjj/juria-blockchain/tests/testutil"
)

type MajorityKeepRunning struct{}

func (expm *MajorityKeepRunning) Name() string {
	return "majority_keep_running"
}

// Keep majority (2f+1) validators running while stopping the rest
// The blockchain should keep remain healthy
// When the stopped nodes up again, they should sync the history
func (expm *MajorityKeepRunning) Run(cls *cluster.Cluster) error {
	total := cls.NodeCount()
	faulty := testutil.PickUniqueRandoms(total, total-core.MajorityCount(total))
	for _, i := range faulty {
		cls.GetNode(i).Stop()
	}
	fmt.Printf("Stopped %d out of %d nodes: %v\n", len(faulty), total, faulty)

	testutil.Sleep(10 * time.Second)
	if err := testutil.HealthCheckMajority(cls); err != nil {
		return err
	}
	for _, fi := range faulty {
		if err := cls.GetNode(fi).Start(); err != nil {
			return err
		}
	}
	fmt.Printf("Started nodes: %v\n", faulty)
	// stopped nodes should sync with the majority after some duration
	testutil.Sleep(10 * time.Second)
	return nil
}
