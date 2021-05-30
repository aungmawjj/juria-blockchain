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
	"github.com/aungmawjj/juria-blockchain/tests/testutil"
)

const (
	WorkDir       = "./workdir"
	NodeCount     = 7
	ClusterDebug  = true
	LoadReqPerSec = 10

	RemoteLinuxCluster = false
)

func setupExperiments() []Experiment {
	expms := make([]Experiment, 0)
	expms = append(expms, &experiments.RestartCluster{})
	expms = append(expms, &experiments.MajorityKeepRunning{})
	return expms
}

func main() {
	cmd := exec.Command("go", "build", "../cmd/juria")
	if RemoteLinuxCluster {
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "GOOS=linux")
		fmt.Printf("\n$ export %s", "GOOS=linux")
	}
	fmt.Printf("\n$ %s\n\n", strings.Join(cmd.Args, " "))
	check(cmd.Run())

	fmt.Println("NodeCount =", NodeCount)
	clustersDir := path.Join(WorkDir, "clusters")
	os.Mkdir(WorkDir, 0755)
	os.Mkdir(clustersDir, 0755)

	cfactory, err := cluster.NewLocalFactory(cluster.LocalFactoryParams{
		JuriaPath: "./juria",
		WorkDir:   clustersDir,
		NodeCount: NodeCount,
		PortN0:    node.DefaultConfig.Port,
		ApiPortN0: node.DefaultConfig.APIPort,
		Debug:     ClusterDebug,
	})
	check(err)

	r := &ExperimentRunner{
		experiments:   setupExperiments(),
		cfactory:      cfactory,
		loadClient:    testutil.NewJuriaCoinLoadClient(100),
		loadReqPerSec: LoadReqPerSec,
	}
	pass, fail := r.run()
	fmt.Printf("\nTotal: %d  |  Pass: %d  |  Fail: %d\n", len(r.experiments), pass, fail)
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
