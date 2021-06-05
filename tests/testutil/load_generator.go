// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package testutil

import (
	"context"
	"net/http"
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
	// to make load test http client efficient
	transport := (http.DefaultTransport.(*http.Transport))
	transport.MaxIdleConns = 100
	transport.MaxIdleConnsPerHost = 100

	return &LoadGenerator{
		txPerSec: tps,
		client:   client,
	}
}

func (lg *LoadGenerator) SetupOnCluster(cls *cluster.Cluster) error {
	return lg.client.SetupOnCluster(cls)
}

func (lg *LoadGenerator) Run(ctx context.Context) {
	jobPerTick := 20
	delay := time.Second / time.Duration(lg.txPerSec/jobPerTick)
	ticker := time.NewTicker(delay)
	defer ticker.Stop()

	jobCh := make(chan struct{}, lg.txPerSec)
	defer close(jobCh)

	for i := 0; i < lg.txPerSec; i++ {
		go lg.loadWorker(jobCh)
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for i := 0; i < jobPerTick; i++ {
				jobCh <- struct{}{}
			}
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

func (lg *LoadGenerator) SetTxPerSec(val int) {
	lg.txPerSec = val
}

func (lg *LoadGenerator) GetTxPerSec() int {
	return lg.txPerSec
}

func (lg *LoadGenerator) GetClient() LoadClient {
	return lg.client
}
