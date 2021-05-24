// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package consensus

import (
	"time"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/hotstuff"
)

type Config struct {
	ChainID int64

	BlockTxLimit int
	TxWaitTime   time.Duration

	BeatDelay     time.Duration
	ViewWidth     time.Duration
	LeaderTimeout time.Duration
}

type Consensus struct {
	resources *Resources

	config Config

	lastBlk   *core.Block
	state     *state
	hsDriver  *hsDriver
	hotstuff  *hotstuff.Hotstuff
	validator *validator
	pacemaker *pacemaker
}

func New(resources *Resources, config Config) *Consensus {
	cons := &Consensus{
		resources: resources,
		config:    config,
	}
	if cons.config.BlockTxLimit == 0 {
		cons.config.BlockTxLimit = 200
	}
	if cons.config.TxWaitTime == 0 {
		cons.config.TxWaitTime = 1 * time.Second
	}
	if cons.config.BeatDelay == 0 {
		cons.config.BeatDelay = 2 * time.Second
	}
	if cons.config.ViewWidth == 0 {
		cons.config.ViewWidth = 30 * time.Second
	}
	if cons.config.LeaderTimeout == 0 {
		cons.config.LeaderTimeout = 10 * time.Second
	}

	cons.start()
	return cons
}

func (cons *Consensus) GetBlock(hash []byte) *core.Block {
	return cons.state.getBlock(hash)
}

func (cons *Consensus) StopPacemaker() {
	cons.pacemaker.stop()
}

func (cons *Consensus) start() {
	lastBlk, err := cons.resources.Storage.GetLastBlock()
	if err != nil {
		lastBlk = cons.makeGenesisBlock()
	}
	cons.lastBlk = lastBlk

	cons.setupState()
	cons.setupHsDriver()
	cons.setupHotstuff()
	cons.setupValidator()
	cons.setupPacemaker()

	cons.validator.start()
	cons.pacemaker.start()
}

func (cons *Consensus) setupState() {
	cons.state = newState(cons.resources)
	cons.state.setBlock(cons.lastBlk)
}

func (cons *Consensus) makeGenesisBlock() *core.Block {
	genesis := &genesis{
		resources: cons.resources,
		chainID:   cons.config.ChainID,
	}
	b0, q0 := genesis.run()
	return b0.SetQuorumCert(q0)
}

func (cons *Consensus) setupHsDriver() {
	cons.hsDriver = &hsDriver{
		resources: cons.resources,
		state:     cons.state,

		txWaitTime:   cons.config.TxWaitTime,
		blockTxLimit: cons.config.BlockTxLimit,
	}
}

func (cons *Consensus) setupHotstuff() {
	cons.hotstuff = hotstuff.New(
		cons.hsDriver,
		newHsBlock(cons.lastBlk, cons.state),
		newHsQC(cons.lastBlk.QuorumCert(), cons.state),
	)
	cons.hsDriver.hotstuff = cons.hotstuff
}

func (cons *Consensus) setupValidator() {
	cons.validator = &validator{
		resources: cons.resources,
		state:     cons.state,
		hotstuff:  cons.hotstuff,
	}
}

func (cons *Consensus) setupPacemaker() {
	cons.pacemaker = &pacemaker{
		resources: cons.resources,
		state:     cons.state,
		hotstuff:  cons.hotstuff,

		beatDelay:     cons.config.BeatDelay,
		viewWidth:     cons.config.ViewWidth,
		leaderTimeout: cons.config.LeaderTimeout,
	}
}
