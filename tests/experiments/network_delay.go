// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package experiments

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/aungmawjj/juria-blockchain/tests/cluster"
	"github.com/aungmawjj/juria-blockchain/tests/health"
	"github.com/aungmawjj/juria-blockchain/tests/testutil"
)

type NetworkDelay struct {
	Delay time.Duration
}

func (expm *NetworkDelay) Name() string {
	return "network_delay_" + expm.Delay.String()
}

func (expm *NetworkDelay) Run(cls *cluster.Cluster) error {
	effects := make([]string, cls.NodeCount())
	for i := 0; i < cls.NodeCount(); i++ {
		delay := expm.Delay + time.Duration(rand.Int63n(int64(expm.Delay)))
		if err := cls.GetNode(i).EffectDelay(delay); err != nil {
			return err
		}
		effects[i] = delay.String()
	}
	defer cls.RemoveEffects()

	fmt.Printf("Added delay %v\n", effects)
	testutil.Sleep(20 * time.Second)
	if err := health.CheckMajorityNodes(cls); err != nil {
		return err
	}

	cls.RemoveEffects()
	fmt.Println("Removed effects")
	testutil.Sleep(10 * time.Second)
	return nil
}
