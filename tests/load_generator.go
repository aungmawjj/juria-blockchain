// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package main

import (
	"context"
	"time"

	"github.com/aungmawjj/juria-blockchain/tests/cluster"
	"github.com/aungmawjj/juria-blockchain/tests/testutil"
)

type LoadGenerator struct {
	txPerSec int
	client   testutil.LoadClient
}

func (lg *LoadGenerator) SetupOnCluster(cls *cluster.Cluster) error {
	return lg.client.SetupOnCluster(cls)
}

func (lg *LoadGenerator) run(ctx context.Context) {
	delay := time.Second / time.Duration(lg.txPerSec)
	ticker := time.NewTicker(delay)
	defer ticker.Stop()

	jobCh := make(chan struct{}, lg.txPerSec)
	defer close(jobCh)

	for i := 0; i < 100; i++ {
		go lg.loadWorker(jobCh)
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			jobCh <- struct{}{}
		}
	}
}

func (lg *LoadGenerator) loadWorker(jobs <-chan struct{}) {
	for range jobs {
		lg.client.SubmitTx()
	}
}
