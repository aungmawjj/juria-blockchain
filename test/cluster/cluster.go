// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package cluster

type Node interface {
	Start() error
	Stop() error
	GetEndpoint() string
}

type Cluster interface {
	Setup() error
	Start() error
	Stop() error
	GetNode(idx int) Node
}

type ClusterFactory interface {
	GetCluster(name string) Cluster
}
