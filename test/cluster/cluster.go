// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package cluster

type Node interface {
	Start() error
	Stop()
	GetEndpoint() string
}

type Cluster interface {
	Start() error
	Stop()
	NodeCount() int
	GetNode(idx int) Node
}
