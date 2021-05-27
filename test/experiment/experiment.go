// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package experiment

import "github.com/aungmawjj/juria-blockchain/test/cluster"

type Experiment interface {
	Name() string
	Run(c *cluster.Cluster) error
}
