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

type RestartMajority struct{}

func (expm *RestartMajority) Name() string {
	return "restart_majority"
}

func (expm *RestartMajority) Run(cls *cluster.Cluster) error {
	total := cls.NodeCount()
	majority := testutil.PickUniqueRandoms(total, core.MajorityCount(total))
	for _, i := range majority {
		cls.GetNode(i).Stop()
	}
	fmt.Printf("Stopped %d out of %d nodes: %v\n", len(majority), total, majority)
	testutil.Sleep(10 * time.Second)
	for _, i := range majority {
		if err := cls.GetNode(i).Start(); err != nil {
			return err
		}
	}
	fmt.Printf("Started nodes: %v\n", majority)
	testutil.Sleep(10 * time.Second)
	if err := testutil.HealthCheckMajority(cls); err != nil {
		return err
	}
	// after majority nodes healthy, minority nodes will auto sync
	time.Sleep(10 * time.Second)
	return nil
}
