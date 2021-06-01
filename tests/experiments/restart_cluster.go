// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package experiments

import (
	"fmt"
	"time"

	"github.com/aungmawjj/juria-blockchain/tests/cluster"
	"github.com/aungmawjj/juria-blockchain/tests/testutil"
)

type RestartCluster struct{}

func (expm *RestartCluster) Name() string {
	return "restart_cluster"
}

func (expm *RestartCluster) Run(cls *cluster.Cluster) error {
	cls.Stop()
	fmt.Println("Stopped cluster")
	testutil.Sleep(10 * time.Second)

	if err := cls.Start(); err != nil {
		return err
	}
	fmt.Println("Restarted cluster")
	testutil.Sleep(20 * time.Second)
	return nil
}
