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
	"time"

	"github.com/aungmawjj/juria-blockchain/node"
	"github.com/aungmawjj/juria-blockchain/tests/cluster"
	"github.com/aungmawjj/juria-blockchain/tests/experiments"
	"github.com/aungmawjj/juria-blockchain/tests/testutil"
)

const (
	WorkDir   = "./workdir"
	NodeCount = 4

	LoadTxPerSec     = 100
	LoadMintAccounts = 100
	LoadDestAccounts = 10000 // increase dest accounts for benchmark

	// Deploy juriacoin chaincode as bincc type (not embeded in juria node)
	JuriaCoinBinCC = false

	// Run tests in remote linux cluster
	// if false it'll use local cluster (running multiple nodes on single local machine)
	RemoteLinuxCluster  = false
	RemoteTemplateSetup = false
	RemoteLoginName     = "ubuntu"
	RemoteKeySSH        = "serverkey"
	RemoteHostsPath     = "hosts"
	RemoteWorkDir       = "/home/ubuntu/juria-tests"
	RemoteNetworkDevice = "ens5"

	// run benchmark, otherwise run experiments
	RunBenchmark      = false
	BenchmarkDuration = 5 * time.Minute
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
	expms = append(expms, &experiments.CorrectExecution{})
	if RemoteLinuxCluster {
		expms = append(expms, &experiments.NetworkDelay{
			Delay: 100 * time.Millisecond,
		})
		expms = append(expms, &experiments.NetworkPacketLoss{
			Percent: 10,
		})
	}
	return expms
}

func main() {
	printVars()
	os.Mkdir(WorkDir, 0755)
	buildJuria()
	loadGen := makeLoadGenerator()
	if RunBenchmark {
		runBenchmark(loadGen)
	} else {
		runExperiments(loadGen)
	}
}

func runBenchmark(loadGen *testutil.LoadGenerator) {
	if !RemoteLinuxCluster {
		fmt.Println("mush run benchmark on remote cluster")
		os.Exit(1)
		return
	}
	bm := &Benchmark{
		workDir:  path.Join(WorkDir, "benchmarks"),
		duration: BenchmarkDuration,
		interval: 5 * time.Second,
		cfactory: makeRemoteClusterFactory(),
		loadGen:  loadGen,
	}
	check(bm.Run())
}

func runExperiments(loadGen *testutil.LoadGenerator) {
	var cfactory cluster.ClusterFactory
	if RemoteLinuxCluster {
		cfactory = makeRemoteClusterFactory()
	} else {
		cfactory = makeLocalClusterFactory()
	}

	r := &ExperimentRunner{
		experiments: setupExperiments(),
		cfactory:    cfactory,
		loadGen:     loadGen,
	}
	pass, fail := r.Run()
	fmt.Printf("\nTotal: %d  |  Pass: %d  |  Fail: %d\n", len(r.experiments), pass, fail)
}

func printVars() {
	fmt.Println()
	fmt.Println("NodeCount =", NodeCount)
	fmt.Println("LoadTxPerSec=", LoadTxPerSec)
	fmt.Println("RemoteCluster =", RemoteLinuxCluster)
	fmt.Println("RunBenchmark=", RunBenchmark)
	fmt.Println()
}

func buildJuria() {
	cmd := exec.Command("go", "build", "../cmd/juria")
	if RemoteLinuxCluster {
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "GOOS=linux")
		fmt.Printf(" $ export %s\n", "GOOS=linux")
	}
	fmt.Printf(" $ %s\n\n", strings.Join(cmd.Args, " "))
	check(cmd.Run())
}

func makeLoadGenerator() *testutil.LoadGenerator {
	var binccPath string
	if JuriaCoinBinCC {
		buildJuriaCoinBinCC()
		binccPath = "./juriacoin"
	}
	fmt.Println("Preparing load generator")
	loadClient := testutil.NewJuriaCoinClient(LoadMintAccounts, LoadDestAccounts, binccPath)
	return testutil.NewLoadGenerator(LoadTxPerSec, loadClient)
}

func buildJuriaCoinBinCC() {
	cmd := exec.Command("go", "build")
	cmd.Args = append(cmd.Args, "-ldflags", "-s -w")
	cmd.Args = append(cmd.Args, "../execution/bincc/juriacoin")
	if RemoteLinuxCluster {
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "GOOS=linux")
		fmt.Printf(" $ export %s\n", "GOOS=linux")
	}
	fmt.Printf(" $ %s\n\n", strings.Join(cmd.Args, " "))
	check(cmd.Run())
}

func makeLocalClusterFactory() *cluster.LocalFactory {
	ftry, err := cluster.NewLocalFactory(cluster.LocalFactoryParams{
		JuriaPath:  "./juria",
		WorkDir:    path.Join(WorkDir, "local-clusters"),
		NodeCount:  NodeCount,
		NodeConfig: getNodeConfig(),
	})
	check(err)
	return ftry
}

func makeRemoteClusterFactory() *cluster.RemoteFactory {
	ftry, err := cluster.NewRemoteFactory(cluster.RemoteFactoryParams{
		JuriaPath:     "./juria",
		WorkDir:       path.Join(WorkDir, "remote-clusters"),
		NodeCount:     NodeCount,
		NodeConfig:    getNodeConfig(),
		LoginName:     RemoteLoginName,
		KeySSH:        RemoteKeySSH,
		HostsPath:     RemoteHostsPath,
		RemoteWorkDir: RemoteWorkDir,
		SetupRequired: RemoteTemplateSetup,
		NetworkDevice: RemoteNetworkDevice,
	})
	check(err)
	return ftry
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
