// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"

	"github.com/aungmawjj/juria-blockchain/node"
	"github.com/aungmawjj/juria-blockchain/test/cluster"
)

const (
	JuriaPath = "./juria"
	WorkDir   = "./workdir"
	NodeCount = 7
)

func main() {
	os.Mkdir(WorkDir, 0755)

	lcc, err := cluster.SetupLocalCluster(cluster.LocalClusterParams{
		JuriaPath: JuriaPath,
		Workdir:   path.Join(WorkDir, "cluster_0"),
		NodeCount: NodeCount,
		PortN0:    node.DefaultConfig.Port,
	})
	check(err)
	err = lcc.Start()
	check(err)

	fmt.Println("started cluster")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Block until a signal is received.
	s := <-c
	fmt.Println("\nGot signal:", s)
	lcc.Stop()
	fmt.Println("stopped cluster")
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
