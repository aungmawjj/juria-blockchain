// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package experiment

import (
	"fmt"
	"time"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/tests/cluster"
	"github.com/aungmawjj/juria-blockchain/tests/testutil"
)

type RestartRandomValidators struct{}

var _ Experiment = (*RestartRandomValidators)(nil)

func (expm *RestartRandomValidators) Name() string {
	return "restart_random_validators"
}

// Stop (f) out of (3f + 1) random validators
// Verify health of remaining validators
// Restart stopping validators
// Wait for 10s to sync
func (expm *RestartRandomValidators) Run(cls *cluster.Cluster) error {
	total := cls.NodeCount()
	faulty := testutil.PickUniqueRandoms(total, total-core.MajorityCount(total))
	for i := range faulty {
		cls.GetNode(faulty[i]).Stop()
	}

	fmt.Printf("Stopped %d out of %d nodes: %v\n", len(faulty), total, faulty)
	if err := testutil.HealthCheckMajority(cls); err != nil {
		return err
	}

	for _, fi := range faulty {
		if err := cls.GetNode(fi).Start(); err != nil {
			return err
		}
	}
	fmt.Printf("Restarted nodes: %v\n", faulty)
	testutil.Sleep(10 * time.Second)
	return nil
}
