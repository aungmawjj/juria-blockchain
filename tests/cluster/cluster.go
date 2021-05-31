// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package cluster

import (
	"github.com/aungmawjj/juria-blockchain/node"
)

type Node interface {
	Start() error
	Stop()
	IsRunning() bool
	GetEndpoint() string
}

type ClusterFactory interface {
	SetupCluster(name string) (*Cluster, error)
}

type Cluster struct {
	nodeConfig node.Config
	nodes      []Node
}

func (cls *Cluster) NodeConfig() node.Config {
	return cls.nodeConfig
}

func (cls *Cluster) Start() error {
	for _, node := range cls.nodes {
		if err := node.Start(); err != nil {
			return err
		}
	}
	return nil
}

func (cls *Cluster) Stop() {
	for _, node := range cls.nodes {
		node.Stop()
	}
}

func (cls *Cluster) NodeCount() int {
	return len(cls.nodes)
}

func (cls *Cluster) GetNode(idx int) Node {
	if idx >= len(cls.nodes) || idx < 0 {
		return nil
	}
	return cls.nodes[idx]
}
