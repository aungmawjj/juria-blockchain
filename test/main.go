// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aungmawjj/juria-blockchain/node"
	"github.com/aungmawjj/juria-blockchain/test/cluster"
	"github.com/aungmawjj/juria-blockchain/test/experiment"
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
	})
	check(err)

	expms := make([]experiment.Experiment, 0)

	expms = append(expms, &experiment.RestartAllNodes{})

	bold := color.New(color.Bold)
	boldRed := color.New(color.Bold, color.FgRed)
	boldGreen := color.New(color.Bold, color.FgGreen)
	for i, expm := range expms {
		bold.Printf("\nExperiment %d. %s\n", i, expm.Name())
		err := runExperiment(cftry, expm)
		if err != nil {
			fmt.Printf("%s\t%s\n", boldRed.Sprint("FAIL"), expm.Name())
			fmt.Printf("%+v\n", err)
		} else {
			fmt.Printf("%s\t%s\n", boldGreen.Sprint("PASS"), expm.Name())
		}
		fmt.Println()
	}
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
