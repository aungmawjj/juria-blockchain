// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package testutil

import (
	"github.com/aungmawjj/juria-blockchain/tests/cluster"
)

func HealthCheckAll(cls *cluster.Cluster) error {
	// TODO: Liveness
	// get status from all nodes, remember heighest bexec
	// get status after (5)s
	// bexec for each node must be higher than previous one

	// TODO: Safety
	// get status from all nodes, get lowest bexec
	// get block by height for bexec, all MerkleRoot must be equal

	// TODO: Rotate
	// get status from all nodes, get leader with highest qc
	// make a timeout channel for (35)s (viewWidth + 5s)
	// get status every 5s
	// leader change on all nodes must occur before timeout

	_, err := GetConsensusStatus(cls.GetNode(0))
	if err != nil {
		return err
	}
	return nil
}

func HealthCheckMajority(cls *cluster.Cluster) error {
	// TODO: Liveness
	// get status from all nodes, remember heighest bexec
	// get status after  (f * leaderTimeout + 5s)
	// bexec for (2f+1) node must be higher than previous one

	// TODO: Safety
	// get status from all nodes, get lowest bexec
	// get block by height for bexec, all MerkleRoot must be equal

	// TODO: Rotate
	// get status from all nodes, get leader with highest qc
	// make a timeout channel for (viewWidth + 5s)
	// get status every 5s
	// leader change on (2f+1) nodes must occur before timeout

	return nil
}
