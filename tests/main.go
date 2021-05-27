// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package main

import (
	"log"
	"os"
	"path"

	"github.com/aungmawjj/juria-blockchain/node"
	"github.com/aungmawjj/juria-blockchain/tests/cluster"
	"github.com/aungmawjj/juria-blockchain/tests/experiment"
)

const (
	JuriaPath = "./juria"
	WorkDir   = "./workdir"
	NodeCount = 7
)

// Running experiments
const (
	RunRestartCluster          = true
	RunRestartRandomValidators = true
)

func main() {
	clustersDir := path.Join(WorkDir, "clusters")

	os.Mkdir(WorkDir, 0755)
	os.Mkdir(clustersDir, 0755)

	cftry, err := cluster.NewLocalFactory(cluster.LocalFactoryParams{
		JuriaPath: JuriaPath,
		WorkDir:   clustersDir,
		NodeCount: NodeCount,
		PortN0:    node.DefaultConfig.Port,
		ApiPortN0: node.DefaultConfig.APIPort,
	})
	check(err)

	runExperiments(cftry, setupExperiments())
}

func setupExperiments() []experiment.Experiment {
	expms := make([]experiment.Experiment, 0)
	if RunRestartCluster {
		expms = append(expms, &experiment.RestartCluster{})
	}
	if RunRestartRandomValidators {
		expms = append(expms, &experiment.RestartRandomValidators{})
	}
	return expms
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
