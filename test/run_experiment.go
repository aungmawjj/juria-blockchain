// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/aungmawjj/juria-blockchain/test/cluster"
	"github.com/aungmawjj/juria-blockchain/test/experiment"
)

func runExperiment(cftry cluster.ClusterFactory, expm experiment.Experiment) error {
	var err error
	cls, err := cftry.GetCluster(expm.Name())
	if err != nil {
		return err
	}
	fmt.Println("Started cluster.")

	done := make(chan struct{})
	go func() {
		defer func() {
			done <- struct{}{}
		}()
		err = cls.Start()
		if err != nil {
			return
		}
		fmt.Println("Started cluster. wait for 10s ...")
		time.Sleep(10 * time.Second)
		// TODO: health check

		fmt.Println("Running experiment")
		err = expm.Run(cls)
		if err != nil {
			return
		}

		// TODO: health check
	}()

	killed := make(chan os.Signal, 1)
	signal.Notify(killed, os.Interrupt)

	select {
	case s := <-killed:
		fmt.Println("\nGot signal:", s)
	case <-done:
	}
	cls.Stop()
	fmt.Println("Stopped cluster.")
	return err
}
