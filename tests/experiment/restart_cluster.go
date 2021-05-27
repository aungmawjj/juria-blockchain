// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package experiment

import (
	"fmt"
	"time"

	"github.com/aungmawjj/juria-blockchain/tests/cluster"
)

type RestartCluster struct{}

var _ Experiment = (*RestartCluster)(nil)

func (expm *RestartCluster) Name() string {
	return "restart_cluster"
}

func (expm *RestartCluster) Run(cls *cluster.Cluster) error {
	cls.Stop()
	fmt.Println("Stopped cluster, wait for 5s")
	time.Sleep(5 * time.Second)

	if err := cls.Start(); err != nil {
		return err
	}
	fmt.Printf("Started cluster, wait for %s\n", cluster.StartCooldown)
	time.Sleep(cluster.StartCooldown)
	return nil
}
