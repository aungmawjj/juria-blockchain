// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"

	"github.com/aungmawjj/juria-blockchain/tests/cluster"
	"github.com/aungmawjj/juria-blockchain/tests/experiment"
	"github.com/aungmawjj/juria-blockchain/tests/testutil"
	"github.com/fatih/color"
)

func runExperiments(cftry cluster.ClusterFactory, expms []experiment.Experiment) {
	bold := color.New(color.Bold)
	boldGrean := color.New(color.Bold, color.FgGreen)
	boldRed := color.New(color.Bold, color.FgRed)

	fmt.Println("\nRunning Experiments")
	for i, expm := range expms {
		bold.Printf("%3d. %s\n", i, expm.Name())
	}

	killed := make(chan os.Signal, 1)
	signal.Notify(killed, os.Interrupt)

	for i, expm := range expms {
		select {
		case <-killed:
			return
		default:
		}
		bold.Printf("\nExperiment %d. %s\n", i, expm.Name())
		err := runSingleExperiment(cftry, expm)
		if err != nil {
			fmt.Printf("%s %s\n", boldRed.Sprint("FAIL"), bold.Sprint(expm.Name()))
			fmt.Printf("error: %+v\n", err)
		} else {
			fmt.Printf("%s %s\n", boldGrean.Sprint("PASS"), bold.Sprint(expm.Name()))
		}
	}
}

func runSingleExperiment(cftry cluster.ClusterFactory, expm experiment.Experiment) error {
	var err error
	cls, err := cftry.SetupCluster(expm.Name())
	if err != nil {
		return err
	}

	done := make(chan struct{})
	go func() {
		defer func() {
			done <- struct{}{}
		}()
		err = cls.Start()
		if err != nil {
			return
		}
		fmt.Println("Started cluster")
		testutil.Sleep(cluster.StartCooldown)
		if err := testutil.HealthCheckAll(cls); err != nil {
			fmt.Printf("health check failed before experiment, %+v\n", err)
			cls.Stop()
			os.Exit(1)
			return
		}

		fmt.Println("==> Running experiment")
		err = expm.Run(cls)
		if err != nil {
			return
		}
		fmt.Println("==> Finished experiment")
		err = testutil.HealthCheckAll(cls)
	}()

	killed := make(chan os.Signal, 1)
	signal.Notify(killed, os.Interrupt)

	select {
	case s := <-killed:
		fmt.Println("\nGot signal:", s)
		err = errors.New("interrupted")
	case <-done:
	}
	cls.Stop()
	fmt.Println("Stopped cluster")
	return err
}
