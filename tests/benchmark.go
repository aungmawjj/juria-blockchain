// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package main

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/aungmawjj/juria-blockchain/consensus"
	"github.com/aungmawjj/juria-blockchain/tests/cluster"
	"github.com/aungmawjj/juria-blockchain/tests/testutil"
	"github.com/aungmawjj/juria-blockchain/txpool"
)

type Measurement struct {
	Timestamp   int64
	TxSubmitted int
	TxCommited  int

	Load       float32 // actual tx sent per sec
	Throughput float32 // tx commited per sec
	Latency    time.Duration

	ConsensusStatus map[int]*consensus.Status
	TxPoolStatus    map[int]*txpool.Status
}

type Benchmark struct {
	workDir  string
	duration time.Duration
	interval time.Duration

	cfactory *cluster.RemoteFactory
	loadGen  *testutil.LoadGenerator

	cluster *cluster.Cluster
	err     error

	measurements     []*Measurement
	lastTxCommitedN0 int
	lastMeasuredTime time.Time

	benchmarkName string
	resultDir     string
}

func (bm *Benchmark) Run() error {
	bm.benchmarkName = fmt.Sprintf("bench_n_%d_load_%d",
		bm.cfactory.GetParams().NodeCount, bm.loadGen.GetTxPerSec())
	if JuriaCoinBinCC {
		bm.benchmarkName += "bincc"
	}

	fmt.Printf("Running benchmark %s\n", bm.benchmarkName)

	bm.resultDir = path.Join(bm.workDir, bm.benchmarkName)
	os.Mkdir(bm.workDir, 0755)
	os.Mkdir(bm.resultDir, 0755)

	done := make(chan struct{})
	loadCtx, stopLoad := context.WithCancel(context.Background())
	go bm.runAsync(loadCtx, done)

	killed := make(chan os.Signal, 1)
	signal.Notify(killed, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-killed:
		fmt.Println("\nGot signal:", s)
		bm.err = errors.New("interrupted")
	case <-done:
	}
	stopLoad()
	if bm.cluster != nil {
		fmt.Println("Stopping cluster")
		bm.cluster.Stop()
		fmt.Println("Stopped cluster")

		bm.stopDstat()
		bm.downDstat()
		fmt.Println("Download dstat records")
	}
	return bm.err
}

func (bm *Benchmark) runAsync(loadCtx context.Context, done chan struct{}) {
	defer close(done)

	fmt.Println("Setting up a new cluster")
	bm.cluster, bm.err = bm.cfactory.SetupCluster(bm.benchmarkName)
	if bm.err != nil {
		return
	}

	if RemoteInstallDstat {
		bm.installDstat()
	}
	bm.startDstat()

	fmt.Println("Starting cluster")
	bm.err = bm.cluster.Start()
	if bm.err != nil {
		return
	}
	fmt.Println("Started cluster")
	testutil.Sleep(10 * time.Second)

	fmt.Println("Setting up load generator")
	bm.err = bm.loadGen.SetupOnCluster(bm.cluster)
	if bm.err != nil {
		return
	}
	go bm.loadGen.Run(loadCtx)
	testutil.Sleep(20 * time.Second)

	bm.err = testutil.HealthCheckAll(bm.cluster)
	if bm.err != nil {
		fmt.Printf("health check failed before benchmark, %+v\n", bm.err)
		bm.cluster.Stop()
		return
	}

	bm.err = bm.measure()
	if bm.err != nil {
		return
	}
}

func (bm *Benchmark) installDstat() {
	var wg sync.WaitGroup
	for i := 0; i < bm.cluster.NodeCount(); i++ {
		node := bm.cluster.GetNode(i).(*cluster.RemoteNode)
		wg.Add(1)
		go func() {
			defer wg.Done()
			node.InstallDstat()
		}()
	}
	wg.Wait()
}

func (bm *Benchmark) startDstat() {
	var wg sync.WaitGroup
	for i := 0; i < bm.cluster.NodeCount(); i++ {
		node := bm.cluster.GetNode(i).(*cluster.RemoteNode)
		wg.Add(1)
		go func() {
			defer wg.Done()
			node.StartDstat()
		}()
	}
	wg.Wait()
}

func (bm *Benchmark) stopDstat() {
	var wg sync.WaitGroup
	for i := 0; i < bm.cluster.NodeCount(); i++ {
		node := bm.cluster.GetNode(i).(*cluster.RemoteNode)
		wg.Add(1)
		go func() {
			defer wg.Done()
			node.StopDstat()
		}()
	}
	wg.Wait()
}

func (bm *Benchmark) downDstat() {
	var wg sync.WaitGroup
	for i := 0; i < bm.cluster.NodeCount(); i++ {
		node := bm.cluster.GetNode(i).(*cluster.RemoteNode)
		filePath := path.Join(bm.resultDir, fmt.Sprintf("dstat_%d.csv", i))
		wg.Add(1)
		go func() {
			defer wg.Done()
			node.DownloadDstat(filePath)
		}()
	}
	wg.Wait()
}

func (bm *Benchmark) measure() error {
	timer := time.NewTimer(bm.duration)
	defer timer.Stop()

	bm.onStartMeasure()

	ticker := time.NewTicker(bm.interval)
	defer ticker.Stop()

	bm.measurements = make([]*Measurement, 0)
	defer bm.saveResults()

	for {
		select {
		case <-timer.C:
			return nil

		case <-ticker.C:
			if err := bm.onTick(); err != nil {
				return err
			}
		}
	}
}

func (bm *Benchmark) onStartMeasure() {
	bm.loadGen.ResetTotalSubmitted()
	consStatus := testutil.GetStatusAll(bm.cluster)
	bm.lastTxCommitedN0 = consStatus[0].CommitedTxCount
	bm.lastMeasuredTime = time.Now()
	fmt.Printf("\nStarted performance measurements\n")
}

func (bm *Benchmark) onTick() error {
	meas := new(Measurement)
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		meas.ConsensusStatus = testutil.GetStatusAll(bm.cluster)
	}()
	go func() {
		defer wg.Done()
		meas.TxPoolStatus = testutil.GetTxPoolStatusAll(bm.cluster)
	}()
	go func() {
		defer wg.Done()
		meas.Latency = bm.measureLatency()
	}()
	wg.Wait()
	meas.Timestamp = time.Now().Unix()
	meas.TxSubmitted = bm.loadGen.ResetTotalSubmitted()

	if len(meas.ConsensusStatus) < bm.cluster.NodeCount() {
		return fmt.Errorf("failed to get consensus status from %d nodes",
			bm.cluster.NodeCount()-len(meas.ConsensusStatus))
	}
	if len(meas.TxPoolStatus) < bm.cluster.NodeCount() {
		return fmt.Errorf("failed to get txpool status from %d nodes",
			bm.cluster.NodeCount()-len(meas.TxPoolStatus))
	}

	meas.TxCommited = meas.ConsensusStatus[0].CommitedTxCount - bm.lastTxCommitedN0
	bm.lastTxCommitedN0 = meas.ConsensusStatus[0].CommitedTxCount

	elapsed := time.Since(bm.lastMeasuredTime)
	bm.lastMeasuredTime = time.Now()
	meas.Load = float32(meas.TxSubmitted) / float32(elapsed.Seconds())
	meas.Throughput = float32(meas.TxCommited) / float32(elapsed.Seconds())

	bm.measurements = append(bm.measurements, meas)

	log.Printf("  Load: %6.1f  |  Throughput: %6.1f  |  Latency: %s\n",
		meas.Load, meas.Throughput, meas.Latency.String())

	return nil
}

func (bm *Benchmark) measureLatency() time.Duration {
	start := time.Now()
	bm.loadGen.GetClient().SubmitTxAndWait()
	return time.Since(start)
}

func (bm *Benchmark) saveResults() error {
	fmt.Println("\nSaving results")
	if err := bm.savePerformance(); err != nil {
		return err
	}
	for i := 0; i < bm.cluster.NodeCount(); i++ {
		if err := bm.saveStatusOneNode(i); err != nil {
			return err
		}
	}
	fmt.Printf("\nSaved Results in %s\n", bm.resultDir)
	return nil
}

func (bm *Benchmark) savePerformance() error {
	f, err := os.Create(path.Join(bm.resultDir, "performance.csv"))
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	w.Write([]string{
		"Timestamp",
		"TxSubmitted",
		"TxCommited",
		"Load",
		"Throughput",
		"Latency",
	})
	var loadTotal, tpTotal float32
	var ltTotal time.Duration
	for _, m := range bm.measurements {
		loadTotal += m.Load
		tpTotal += m.Throughput
		ltTotal += m.Latency

		w.Write([]string{
			strconv.Itoa(int(m.Timestamp)),
			strconv.Itoa(int(m.TxSubmitted)),
			strconv.Itoa(int(m.TxCommited)),
			fmt.Sprintf("%.2f", m.Load),
			fmt.Sprintf("%.2f", m.Throughput),
			m.Latency.String(),
		})
	}

	loadAvg := loadTotal / float32(len(bm.measurements))
	tpAvg := tpTotal / float32(len(bm.measurements))
	ltAvg := ltTotal / time.Duration(len(bm.measurements))

	fmt.Println("\nAverage:")
	log.Printf("  Load: %6.1f  |  Throughput: %6.1f  |  Latency: %s\n",
		loadAvg, tpAvg, ltAvg.String())

	w.Write([]string{
		strconv.Itoa(int(time.Now().Unix())), "", "",
		fmt.Sprintf("%.2f", loadAvg),
		fmt.Sprintf("%.2f", tpAvg),
		ltAvg.String(),
	})
	w.Flush()
	return nil
}

func (bm *Benchmark) saveStatusOneNode(i int) error {
	f, err := os.Create(path.Join(bm.resultDir, fmt.Sprintf("status_%d.csv", i)))
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	w.Write([]string{
		"Timestamp",
		"BlockPoolSize",
		"QCPoolSize",
		"LeaderIndex",
		"ViewStart",
		"QCHigh",
		"TxPoolTotal",
		"TxPoolPending",
		"TxPoolQueue",
	})
	for _, m := range bm.measurements {
		w.Write([]string{
			strconv.Itoa(int(m.Timestamp)),
			strconv.Itoa(m.ConsensusStatus[i].BlockPoolSize),
			strconv.Itoa(m.ConsensusStatus[i].QCPoolSize),
			strconv.Itoa(int(m.ConsensusStatus[i].LeaderIndex)),
			strconv.Itoa(int(m.ConsensusStatus[i].ViewStart)),
			strconv.Itoa(int(m.ConsensusStatus[i].QCHigh)),

			strconv.Itoa(m.TxPoolStatus[i].Total),
			strconv.Itoa(m.TxPoolStatus[i].Pending),
			strconv.Itoa(m.TxPoolStatus[i].Queue),
		})
	}
	w.Flush()
	return nil
}
