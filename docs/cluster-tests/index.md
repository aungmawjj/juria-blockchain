---
layout: default
title: Cluster Tests
nav_order: 3
---

# Cluster Tests
{: .no_toc}
Cluster tests introduce different failures to the system and verify that safety is not breached during those experiments.
The structure of the tests is inspired by [Diem](https://github.com/diem/diem) cluster tests.
You can run the tests on both local machine and remote linux servers.
Implementation can be found [here](https://github.com/aungmawjj/juria-blockchain/tree/master/tests).

## Table of contents
{: .no_toc .text-delta}

* TOC
{:toc}

## Structure
Major components of tests include:

**Experiment** is some condition we want to test our system with such as restart some nodes.

**HealthCheck** is how we verify whether the system is running correctly or not. 
Runner verifies that cluster is healthy before and after running each experiment.
Health is verified in three aspects:
1. **Safety**: nodes must have the same state at a certain block height
2. **Liveness**:  the system must commits blocks and transactions
3. **Leader Rotation**: the system must rotate leader

## Setup Development Environment

1. Install dependencies
```bash
# MacOS
xcode-select --install
```
```bash
# Ubuntu
sudo apt-get install build-essential
```
2. Download and install [`go 1.16`](https://golang.org/doc/install)
3. Prepare the repo
```bash
git clone https://github.com/aungmawjj/juria-blockchain
cd juria-blockchain
go mod tidy
```

## Local Cluster
Run tests on local cluster.
```bash
cd tests
go run .
```

The test script will compile `juria` and set up a cluster of 4 nodes with different ports on the local machine.
Runner will run Experiments and perform HealthChecks.

***NOTE**: Network simulation experiments are only run on the remote linux cluster.*

## Remote Linux Cluster

Configure `main.go` to run tests on remote linux servers.
```go
RemoteLinuxCluster = true // set true
```
Provide two files:
1. `serverkey` : ssh key for remote login
2. `hosts` : hostnames (ip addresses) with each line for one server

Run tests.
```bash
cd tests
go run .
```
