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
	BeatTimeout time.Duration

	// minimum delay between each block (i.e, it can define maximum block rate)
	BlockDelay time.Duration

	// view duration for a leader
	ViewWidth time.Duration
}

var DefaultConfig = Config{
	BlockTxLimit: 500,
	TxWaitTime:   1 * time.Second,
	BeatTimeout:  500 * time.Millisecond,
	BlockDelay:   50 * time.Millisecond, // maximum block rate = 20 blk per sec
	ViewWidth:    30 * time.Second,
}
