---
layout: default
title: Benchmark
nav_order: 4
---

# Benchmark
{: .no_toc}
Throughput and latency for different cluster sizes are measured
using benchmark script in `tests` directory.

## Table of contents
{: .no_toc .text-delta}

* TOC
{:toc}

## Running benchmark
[Setup development environment]({{site.baseurl}}{% link cluster-tests/index.md %}/#setup-development-environment).

Provide two files in `tests` directory:
1. `serverkey` : ssh key for remote login
2. `hosts` : hostnames (ip addresses) with each line for one server

Configure `main.go` to run benchmarks on remote linux servers.
```go
RemoteLinuxCluster = true // set true
RunBenchmark       = true // set true 
```

Run benchmarks.
```bash
go run .
```

## Benchmark on AWS
Benchmarks using different cluster sizes are run on AWS.
The results, logs, and resource utilization (dstat) can be found [here](https://drive.google.com/drive/folders/1ob9hn_B7JTRdwPUoQh2psqFHjyx4YgcW?usp=sharing).
![Benchmark](/assets/images/benchmark_juria.png)
