// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package testutil

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/aungmawjj/juria-blockchain/tests/cluster"
)

type LoadGenerator struct {
	txPerSec int
	client   LoadClient

	totalSubmitted int64
}

func NewLoadGenerator(tps int, client LoadClient) *LoadGenerator {
	return &LoadGenerator{
		txPerSec: tps,
		client:   client,
	}
}

func (lg *LoadGenerator) SetupOnCluster(cls *cluster.Cluster) error {
	return lg.client.SetupOnCluster(cls)
}

func (lg *LoadGenerator) Run(ctx context.Context) {
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
		if _, _, err := lg.client.SubmitTx(); err == nil {
			lg.increaseSubmitted()
		}
	}
}

func (lg *LoadGenerator) increaseSubmitted() {
	atomic.AddInt64(&lg.totalSubmitted, 1)
}

func (lg *LoadGenerator) ResetTotalSubmitted() int {
	return int(atomic.SwapInt64(&lg.totalSubmitted, 0))
}

func (lg *LoadGenerator) GetTxPerSec() int {
	return lg.txPerSec
}

func (lg *LoadGenerator) GetClient() LoadClient {
	return lg.client
}
