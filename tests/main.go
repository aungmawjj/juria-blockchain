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
	NodeCount     = 4
	LoadReqPerSec = 100

	RemoteLinuxCluster = false
	RemoteSendJuria    = false
	RemoteLoginName    = "ubuntu"
	RemoteKeySSH       = "serverkey"
	RemoteHostsPath    = "hosts"
	RemoteWorkDir      = "/home/ubuntu/juria-tests"
)

func getNodeConfig() node.Config {
	config := node.DefaultConfig
	config.Debug = true
	return config
}

func setupExperiments() []Experiment {
	expms := make([]Experiment, 0)
	expms = append(expms, &experiments.RestartCluster{})
	expms = append(expms, &experiments.MajorityKeepRunning{})
	return expms
}

func main() {
	fmt.Println()
	fmt.Println("NodeCount =", NodeCount)
	fmt.Println("LoadReqPerSec =", LoadReqPerSec)
	fmt.Println("RemoteCluster =", RemoteLinuxCluster)
	fmt.Println()

	cmd := exec.Command("go", "build", "../cmd/juria")
	if RemoteLinuxCluster {
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "GOOS=linux")
		fmt.Printf(" $ export %s\n", "GOOS=linux")
	}
	fmt.Printf(" $ %s\n\n", strings.Join(cmd.Args, " "))
	check(cmd.Run())

	os.Mkdir(WorkDir, 0755)

	var cfactory cluster.ClusterFactory
	if RemoteLinuxCluster {
		cfactory = makeRemoteClusterFactory()
	} else {
		cfactory = makeLocalClusterFactory()
	}

	r := &ExperimentRunner{
		experiments:   setupExperiments(),
		cfactory:      cfactory,
		loadClient:    testutil.NewJuriaCoinLoadClient(LoadReqPerSec * 2),
		loadReqPerSec: LoadReqPerSec,
	}
	pass, fail := r.run()
	fmt.Printf("\nTotal: %d  |  Pass: %d  |  Fail: %d\n", len(r.experiments), pass, fail)
}

func makeLocalClusterFactory() cluster.ClusterFactory {
	clustersDir := path.Join(WorkDir, "local-clusters")
	os.Mkdir(clustersDir, 0755)

	ftry, err := cluster.NewLocalFactory(cluster.LocalFactoryParams{
		JuriaPath:  "./juria",
		WorkDir:    clustersDir,
		NodeCount:  NodeCount,
		NodeConfig: getNodeConfig(),
	})
	check(err)
	return ftry
}

func makeRemoteClusterFactory() cluster.ClusterFactory {
	clustersDir := path.Join(WorkDir, "remote-clusters")
	os.Mkdir(clustersDir, 0755)

	ftry, err := cluster.NewRemoteFactory(cluster.RemoteFactoryParams{
		JuriaPath:     "./juria",
		WorkDir:       clustersDir,
		NodeCount:     NodeCount,
		NodeConfig:    getNodeConfig(),
		LoginName:     RemoteLoginName,
		KeySSH:        RemoteKeySSH,
		HostsPath:     RemoteHostsPath,
		RemoteWorkDir: RemoteWorkDir,
	})
	check(err)
	return ftry
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
