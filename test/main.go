// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/aungmawjj/juria-blockchain/node"
	"github.com/aungmawjj/juria-blockchain/test/cluster"
)

const (
	Juria     = "./juria"
	WorkDir   = "workdir"
	NodeCount = 4
)

func main() {
	os.RemoveAll(WorkDir)
	err := os.Mkdir(WorkDir, 0755)
	check(err)

	ftry := &cluster.LocalClusterFactory{
		JuriaPath: Juria,
		Workdir:   WorkDir,
		NodeCount: NodeCount,
		PortN0:    node.DefaultConfig.Port,
	}
	cluster := ftry.GetCluster("cluster_1")
	err = cluster.Setup()
	check(err)
	err = cluster.Start()
	check(err)

	fmt.Println("started cluster")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Block until a signal is received.
	s := <-c
	fmt.Println("\nGot signal:", s)
	cluster.Stop()
	fmt.Println("stopped cluster")
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
