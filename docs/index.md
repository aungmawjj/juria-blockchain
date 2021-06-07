---
layout: default
title: Home
nav_order: 1
description: High-performance consortium blockchain
---

# Juria Blockchain
Juria is a high-performance consortium blockchain with [Hotstuff](https://arxiv.org/abs/1803.05069) consensus mechanism and a transaction-based state machine.

Hotstuff provides a mechanism to rotate leader (block maker) efficiently among the validator nodes. Hence it is not required to have a single trusted leader in the network.

With the use of three-chain Hotstuff commit rule, Juria ensures that the same history of blocks are commited on all nodes despite network and machine failures.

[Get started now](#getting-started){: .btn .btn-primary .fs-5 .mb-4 .mb-md-0 .mr-2 }
[View it on GitHub](https://github.com/aungmawjj/juria-blockchain){: .btn .fs-5 .mb-4 .mb-md-0 }

![Benchmark](assets/images/benchmark_juria.png)

## Getting started
You can run the cluster tests on local machine in a few seconds.
`go 1.16` is required.

1. Prepare the repo
```sh
git clone https://github.com/aungmawjj/juria-blockchain
cd juria-blockchain
go mod tidy
```

2. Run tests
```sh
cd tests
go run .
```
The test script will compile `cmd/juria` and setup cluster of 4 nodes with different ports on local machine.
Experiments from `tests/experiments` will be run and health check will be performed throughout the tests.

***NOTE**: Network simmulation experiments are only run on the remote linux cluster.*

## Documentation
* [Key Concepts]({{site.baseurl}}{% link key-concepts/index.md %})
* [Tests and Evaluation]({{site.baseurl}}{% link tests-and-evaluation/index.md %})
* [Benchmark on AWS]({{site.baseurl}}{% link benchmark-on-aws/index.md %})
* [Setup Juria Cluster]({{site.baseurl}}{% link setup-juria-cluster/index.md %})

## About the project
### License
This project is licensed under the GPL-3.0 License.

### Contributing
When contributing to this repository, please first discuss the change you wish to make via issue, email, or any other method with the owners of this repository before making a change.
