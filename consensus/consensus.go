// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package consensus

import (
	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/hotstuff"
)

type Consensus struct {
	resources *Resources

	config Config

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
	return cons
}

func (cons *Consensus) Start() {
	cons.start()
}

func (cons *Consensus) Stop() {
	cons.stop()
}

func (cons *Consensus) GetStatus() Status {
	return cons.getStatus()
}

func (cons *Consensus) GetBlock(hash []byte) *core.Block {
	return cons.state.getBlock(hash)
}

func (cons *Consensus) start() {
	lastBlk, err := cons.resources.Storage.GetLastBlock()
	if err != nil {
		lastBlk = cons.makeGenesisBlock()
	}

	cons.setupState(lastBlk)
	cons.setupHsDriver()
	cons.setupHotstuff(lastBlk)
	cons.setupValidator()
	cons.setupPacemaker()

	cons.validator.start()
	cons.pacemaker.start()
}

func (cons *Consensus) stop() {
	if cons.pacemaker == nil {
		return
	}
	cons.pacemaker.stop()
	cons.validator.stop()
}

func (cons *Consensus) setupState(lastBlk *core.Block) {
	cons.state = newState(cons.resources)
	cons.state.setBlock(lastBlk)
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
		config:    cons.config,
		state:     cons.state,
	}
}

func (cons *Consensus) setupHotstuff(lastBlk *core.Block) {
	cons.hotstuff = hotstuff.New(
		cons.hsDriver,
		newHsBlock(lastBlk, cons.state),
		newHsQC(lastBlk.QuorumCert(), cons.state),
	)
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
		config:    cons.config,
		state:     cons.state,
		hotstuff:  cons.hotstuff,
	}
}

func (cons *Consensus) getStatus() (status Status) {
	if cons.pacemaker == nil {
		return status
	}
	status.Started = true
	status.BlockPoolSize = cons.state.getBlockPoolSize()
	status.LeaderIndex = cons.state.getLeaderIndex()
	status.ViewStart = cons.pacemaker.getViewStart()
	status.PendingViewChange = cons.pacemaker.getPendingViewChange()

	status.BVote = cons.hotstuff.GetBVote().Height()
	status.BLeaf = cons.hotstuff.GetBLeaf().Height()
	status.BLock = cons.hotstuff.GetBLock().Height()
	status.BExec = cons.hotstuff.GetBExec().Height()
	qcHighRef := cons.hotstuff.GetQCHigh().Block()
	if qcHighRef != nil {
		status.QCHigh = qcHighRef.Height()
	}
	return status
}
