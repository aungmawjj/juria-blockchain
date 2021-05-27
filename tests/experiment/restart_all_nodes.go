// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package experiment

import (
	"fmt"
	"time"

	"github.com/aungmawjj/juria-blockchain/tests/cluster"
)

type RestartAllNodes struct{}

var _ Experiment = (*RestartAllNodes)(nil)

func (expm *RestartAllNodes) Name() string {
	return "restart_all_nodes"
}

func (expm *RestartAllNodes) Run(cls *cluster.Cluster) error {
	fmt.Println("Stopping cluster")

	cls.Stop()
	time.Sleep(5 * time.Second)
	fmt.Println("Stopped cluster")

	if err := cls.Start(); err != nil {
		return err
	}
	fmt.Printf("Started cluster, wait for %s", cluster.StartCooldown)
	time.Sleep(cluster.StartCooldown)
	return nil
}