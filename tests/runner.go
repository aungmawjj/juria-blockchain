// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aungmawjj/juria-blockchain/tests/cluster"
	"github.com/aungmawjj/juria-blockchain/tests/testutil"
	"github.com/fatih/color"
)

type Experiment interface {
	Name() string
	Run(c *cluster.Cluster) error
}

type ExperimentRunner struct {
	experiments []Experiment
	cfactory    cluster.ClusterFactory

	loadReqPerSec int
	loadClient    testutil.LoadClient

	loadJobCh chan struct{}
}

func (r *ExperimentRunner) run() (pass, fail int) {

	r.loadJobCh = make(chan struct{}, 20)
	for i := 0; i < 30; i++ {
		go r.loadWorker(r.loadJobCh)
	}

	bold := color.New(color.Bold)
	boldGrean := color.New(color.Bold, color.FgGreen)
	boldRed := color.New(color.Bold, color.FgRed)

	fmt.Println("\nRunning Experiments")
	for i, expm := range r.experiments {
		bold.Printf("%3d. %s\n", i, expm.Name())
	}

	killed := make(chan os.Signal, 1)
	signal.Notify(killed, os.Interrupt)

	for i, expm := range r.experiments {
		bold.Printf("\nExperiment %d. %s\n", i, expm.Name())
		err := r.runSingleExperiment(expm)
		if err != nil {
			fail++
			fmt.Printf("%s %s\n", boldRed.Sprint("FAIL"), bold.Sprint(expm.Name()))
			fmt.Printf("error: %+v\n", err)
		} else {
			pass++
			fmt.Printf("%s %s\n", boldGrean.Sprint("PASS"), bold.Sprint(expm.Name()))
		}
		select {
		case <-killed:
			return pass, fail
		default:
		}
	}
	return pass, fail
}

func (r *ExperimentRunner) runSingleExperiment(expm Experiment) error {
	var err error
	cls, err := r.cfactory.SetupCluster(expm.Name())
	if err != nil {
		return err
	}

	done := make(chan struct{})
	loadCtx, stopLoad := context.WithCancel(context.Background())
	go func() {
		defer close(done)
		cls.Stop() // to make sure no existing process keeps running
		err = cls.Start()
		if err != nil {
			return
		}
		fmt.Println("Started cluster")
		testutil.Sleep(10 * time.Second)

		fmt.Println("Setting up load generator")
		err = r.loadClient.SetupOnCluster(cls)
		if err != nil {
			return
		}
		go r.runLoadGenerator(loadCtx)

		testutil.Sleep(20 * time.Second)

		if err := testutil.HealthCheckAll(cls); err != nil {
			fmt.Printf("health check failed before experiment, %+v\n", err)
			cls.Stop()
			os.Exit(1)
			return
		}

		fmt.Println("==> Running experiment")
		err = expm.Run(cls)
		if err != nil {
			fmt.Println("==> Experiment failed")
			return
		}
		fmt.Println("==> Finished experiment")
		err = testutil.HealthCheckAll(cls)
	}()

	killed := make(chan os.Signal, 1)
	signal.Notify(killed, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-killed:
		fmt.Println("\nGot signal:", s)
		err = errors.New("interrupted")
	case <-done:
	}
	stopLoad()
	fmt.Println("Stopping cluster")
	cls.Stop()
	fmt.Println("Stopped cluster")
	return err
}

func (r *ExperimentRunner) runLoadGenerator(ctx context.Context) {
	ticker := time.NewTicker(time.Second / time.Duration(r.loadReqPerSec))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.loadJobCh <- struct{}{}
		}
	}
}

func (r *ExperimentRunner) loadWorker(jobs <-chan struct{}) {
	for range jobs {
		r.loadClient.SubmitTx()
	}
}
