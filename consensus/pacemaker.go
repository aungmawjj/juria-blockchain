// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package consensus

import (
	"time"

	"github.com/aungmawjj/juria-blockchain/hotstuff"
	"github.com/aungmawjj/juria-blockchain/logger"
)

type pacemaker struct {
	resources *Resources
	config    Config

	state    *state
	hotstuff *hotstuff.Hotstuff

	stopCh chan struct{}
}

func (pm *pacemaker) start() {
	if pm.stopCh != nil {
		return
	}
	pm.stopCh = make(chan struct{})
	go pm.run()
	logger.I().Info("started pacemaker")
}

func (pm *pacemaker) stop() {
	if pm.stopCh == nil {
		return // not started yet
	}
	select {
	case <-pm.stopCh: // already stopped
		return
	default:
	}
	close(pm.stopCh)
	logger.I().Info("stopped pacemaker")
	pm.stopCh = nil
}

func (pm *pacemaker) run() {
	for {
		blkDelay := time.After(pm.config.BlockDelay)

		pm.onBeat()

		beatWait := pm.config.BeatDelay
		if pm.resources.TxPool.GetStatus().Total == 0 {
			beatWait += pm.config.TxWaitTime
		}
		beatT := time.NewTimer(beatWait)
		subQC := pm.hotstuff.SubscribeNewQCHigh()

		select {
		case <-pm.stopCh:
			return

		// either beatdelay timeout or I'm able to create qc
		case <-beatT.C:
		case <-subQC.Events():
		}
		beatT.Stop()
		subQC.Unsubscribe()

		select {
		case <-pm.stopCh:
			return
		case <-blkDelay:
		}
	}
}

func (pm *pacemaker) onBeat() {
	pm.state.mtxUpdate.Lock()
	defer pm.state.mtxUpdate.Unlock()

	select {
	case <-pm.stopCh:
		return
	default:
	}
	if !pm.state.isThisNodeLeader() {
		return
	}
	pm.propose()
}

func (pm *pacemaker) propose() {
	blk := pm.hotstuff.OnPropose()
	logger.I().Debugw("proposed block", "height", blk.Height(), "qc", qcRefHeight(blk.Justify()))
	vote := blk.(*hsBlock).block.ProposerVote()
	pm.hotstuff.OnReceiveVote(newHsVote(vote, pm.state))
	pm.hotstuff.Update(blk)
}
