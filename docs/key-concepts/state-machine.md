---
layout: default
title: State Machine
parent: Key Concepts
nav_order: 1
---

# State Machine
{: .no_toc}
Juria blockchain can be represented as a state machine replicated on different computers.

## Table of contents
{: .no_toc .text-delta}

* TOC
{:toc}

## State Merkle Tree
In Juria, the state machine records the current state as key-value pairs.
It is essential to ensure that state machines on different computers reach the same state
after processing a list of inputs.

One way to achieve this is to compute the cryptographic hash of all key-value pairs
and compare the hash among different computers.
However, the downside of this approach is that it recomputes the hash of the whole state 
even for a single value change.

Juria blockchain uses a [Merkle tree](https://en.wikipedia.org/wiki/Merkle_tree) to compute the hash of all key-value pairs efficiently.
A Merkle tree or hash tree is a tree data structure in which each non-leaf node is the cryptographic hash of its child nodes.
The leaf nodes are the cryptographic hash of the corresponding key-value pairs.
The Merkle tree allows efficient and secure verification of the contents of large data structures.

## Chaincode
A chaincode or smart contract is a program with a unique address and isolated state inside the state machine.

When the state machine processes an input, it executes the chaincode at the given address.
During the execution, the chaincode changes its key-value pairs.

## Transaction
A transaction is a unique input with a digital signature from the end-user.
Transactions are the inputs for the state machine which are processed in the same order on different computers.
