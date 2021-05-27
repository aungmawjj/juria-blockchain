// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aungmawjj/juria-blockchain/node"
	"github.com/aungmawjj/juria-blockchain/tests/cluster"
	"github.com/aungmawjj/juria-blockchain/tests/experiment"
	"github.com/fatih/color"
)

var (
	JuriaPath = "./juria"
	WorkDir   = "./workdir"
	NodeCount = 7
)

func main() {
	os.Mkdir(WorkDir, 0755)

	cftry, err := cluster.NewLocalFactory(cluster.LocalFactoryParams{
		JuriaPath: JuriaPath,
		WorkDir:   WorkDir,
		NodeCount: NodeCount,
		PortN0:    node.DefaultConfig.Port,
		ApiPortN0: node.DefaultConfig.APIPort,
	})
	check(err)

	expms := make([]experiment.Experiment, 0)

	expms = append(expms, &experiment.RestartAllNodes{})

	bold := color.New(color.Bold)

	fmt.Printf("Running Experiments. Total: %d\n", len(expms))
	for i, expm := range expms {
		bold.Printf("%3d. %s\n", i, expm.Name())
	}

	for i, expm := range expms {
		bold.Printf("\nExperiment %d. %s\n", i, expm.Name())
		err := runExperiment(cftry, expm)
		if err != nil {
			bold.Printf("%s %s\n", color.RedString("FAIL"), expm.Name())
			fmt.Printf("%+v\n", err)
		} else {
			bold.Printf("%s %s\n", color.GreenString("PASS"), expm.Name())
		}
		fmt.Println()
	}
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
