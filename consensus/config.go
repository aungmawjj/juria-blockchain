// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package consensus

import "time"

type Config struct {
	ChainID int64

	// maximum tx count in a block
	BlockTxLimit int

	// block creation delay if no transactions in the pool
	TxWaitTime time.Duration

	// for leader, delay to propose next block if she cannot create qc")
	BeatDelay time.Duration

	// view duration for a leader
	ViewWidth time.Duration

	// if leader cannot create next qc in this duration, change view
	LeaderTimeout time.Duration
}

var DefaultConfig = Config{
	BlockTxLimit:  200,
	TxWaitTime:    1 * time.Second,
	BeatDelay:     500 * time.Millisecond,
	ViewWidth:     30 * time.Second,
	LeaderTimeout: 6 * time.Second,
}
