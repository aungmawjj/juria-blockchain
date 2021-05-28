// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package experiments

import (
	"fmt"
	"time"

	"github.com/aungmawjj/juria-blockchain/tests/cluster"
	"github.com/aungmawjj/juria-blockchain/tests/testutil"
)

type RestartRandomValidators struct{}

func (expm *RestartRandomValidators) Name() string {
	return "restart_random_validators"
}

func (expm *RestartRandomValidators) Run(cls *cluster.Cluster) error {
	total := cls.NodeCount()
	faulty := testutil.PickUniqueRandoms(total, total/2)
	for i := range faulty {
		cls.GetNode(faulty[i]).Stop()
	}

	fmt.Printf("Stopped %d out of %d nodes: %v\n", len(faulty), total, faulty)
	testutil.Sleep(10 * time.Second)
	for _, fi := range faulty {
		if err := cls.GetNode(fi).Start(); err != nil {
			return err
		}
	}
	fmt.Printf("Started nodes: %v\n", faulty)
	testutil.Sleep(10 * time.Second)
	return nil
}
