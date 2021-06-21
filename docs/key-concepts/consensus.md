---
layout: default
title: Consensus
parent: Key Concepts
nav_order: 2
---

# Consensus
{: .no_toc}
Juria blockchain uses [Hotstuff](https://arxiv.org/abs/1803.05069) BFT consensus mechanism with round-robin leader rotation.

## Table of contents
{: .no_toc .text-delta}

* TOC
{:toc}

## Block
[Transactions]({{site.baseurl}}{% link key-concepts/state-machine.md %}/#transaction) are grouped as order lists using blocks.
When a block is committed, its transactions are processed by the [state machine]({{site.baseurl}}{% link key-concepts/state-machine.md %}).
Blocks are committed by using the Hotstuff three-chain rule.
Hotstuff core implementation can be found [here](https://github.com/aungmawjj/juria-blockchain/blob/master/hotstuff/hotstuff.go).

## Quorum Certificate
A Quorum Certificate (*QC*) is the combination of signatures from majority nodes for a block proposal. 

At any given time, the current leader creates a new block proposal using the latest *QC* and broadcasts it to all validators.
The validators verify the proposal and send the votes with their signatures to the proposer.
The leader creates a new *QC* once the votes from majority nodes are collected.
This way, the leader creates new blocks repeatedly.

## Voting Rules
For voting a proposal, in addition to Hotstuff’s safety and liveness rules, the following rules are checked to ensure the correct state machine synchronization.

1. The last committed block height must be the same.
2. Merkle root of the [state tree]({{site.baseurl}}{% link key-concepts/state-machine.md %}/#state-merkle-tree) must be the same.
3. Included transactions must be valid.

## Leader Rotation
Juria blockchain performs the leader rotation in a round-robin manner.
Every node keeps the list of validators in the same order.
A view is a period in which a selected leader can propose blocks.
Each node holds a timer for the current view.
At the end of the current view, the next leader is selected.
By using Hotstuff, the leader rotation can be performed frequently and efficiently.
