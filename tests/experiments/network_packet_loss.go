// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package experiments

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/aungmawjj/juria-blockchain/tests/cluster"
	"github.com/aungmawjj/juria-blockchain/tests/testutil"
)

type NetworkPacketLoss struct {
	Percent float32
}

func (expm *NetworkPacketLoss) Name() string {
	return fmt.Sprintf("network_packet_loss_%.2f", expm.Percent)
}

func (expm *NetworkPacketLoss) Run(cls *cluster.Cluster) error {
	effects := make([]string, cls.NodeCount())
	for i := 0; i < cls.NodeCount(); i++ {
		percent := expm.Percent + rand.Float32()
		if err := cls.GetNode(i).EffectLoss(percent); err != nil {
			return err
		}
		effects[i] = fmt.Sprintf("%.2f%%", percent)
	}
	defer cls.RemoveEffects()

	fmt.Printf("Added packet loss %v\n", effects)
	testutil.Sleep(20 * time.Second)
	if err := testutil.HealthCheckMajority(cls); err != nil {
		return err
	}

	cls.RemoveEffects()
	fmt.Println("Removed effects")
	testutil.Sleep(10 * time.Second)
	return nil
}
