// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package cluster

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
	nodes []Node
}

func (lcc *Cluster) Start() error {
	for _, node := range lcc.nodes {
		if err := node.Start(); err != nil {
			return err
		}
	}
	return nil
}

func (lcc *Cluster) Stop() {
	for _, node := range lcc.nodes {
		node.Stop()
	}
}

func (lcc *Cluster) NodeCount() int {
	return len(lcc.nodes)
}

func (lcc *Cluster) GetNode(idx int) Node {
	if idx >= len(lcc.nodes) || idx < 0 {
		return nil
	}
	return lcc.nodes[idx]
}
