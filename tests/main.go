// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/aungmawjj/juria-blockchain/node"
	"github.com/aungmawjj/juria-blockchain/tests/cluster"
	"github.com/aungmawjj/juria-blockchain/tests/experiments"
)

const (
	WorkDir      = "./workdir"
	NodeCount    = 7
	ClusterDebug = true
)

func setupExperiments() []Experiment {
	expms := make([]Experiment, 0)
	expms = append(expms, &experiments.RestartCluster{})
	expms = append(expms, &experiments.MajorityKeepRunning{})
	expms = append(expms, &experiments.RestartMajority{})
	return expms
}

func main() {
	cmd := exec.Command("go", "build", "../cmd/juria")
	fmt.Printf("\n$ %s\n\n", strings.Join(cmd.Args, " "))
	check(cmd.Run())

	fmt.Println("NodeCount =", NodeCount)
	clustersDir := path.Join(WorkDir, "clusters")
	os.Mkdir(WorkDir, 0755)
	os.Mkdir(clustersDir, 0755)

	cftry, err := cluster.NewLocalFactory(cluster.LocalFactoryParams{
		JuriaPath: "./juria",
		WorkDir:   clustersDir,
		NodeCount: NodeCount,
		PortN0:    node.DefaultConfig.Port,
		ApiPortN0: node.DefaultConfig.APIPort,
		Debug:     ClusterDebug,
	})
	check(err)

	expms := setupExperiments()
	pass, fail := runExperiments(cftry, expms)
	fmt.Printf("\nTotal: %d  |  Pass: %d  |  Fail: %d\n", len(expms), pass, fail)
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
