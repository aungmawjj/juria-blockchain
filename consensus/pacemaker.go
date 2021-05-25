// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package consensus

import (
	"sync"
	"time"

	"github.com/aungmawjj/juria-blockchain/hotstuff"
	"github.com/aungmawjj/juria-blockchain/logger"
)

type pacemaker struct {
	resources *Resources
	config    Config

	state    *state
	hotstuff *hotstuff.Hotstuff

	// start timestamp in second of current view
	viewStart int64
	mtxVS     sync.RWMutex

	// true when view changed before the next leader is approved
	pendingViewChange bool
	mtxPVC            sync.RWMutex

	stopCh chan struct{}
}

func (pm *pacemaker) start() {
	if pm.stopCh != nil {
		return
	}
	pm.stopCh = make(chan struct{})
	pm.setViewStart()
	pm.setPendingViewChange(false)
	go pm.beatLoop()
	go pm.viewChangeLoop()
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
	pm.stopCh = nil
}

func (pm *pacemaker) beatLoop() {
	subQC := pm.hotstuff.SubscribeNewQCHigh()
	defer subQC.Unsubscribe()

	for {
		pm.onBeat()
		d := pm.config.BeatDelay
		if pm.resources.TxPool.GetStatus().Total == 0 {
			d += pm.config.TxWaitTime
		}
		select {
		case <-pm.stopCh:
			return
		case <-time.After(d):
		case <-subQC.Events():
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
	hsBlk := pm.hotstuff.OnPropose()
	pm.logProposal(hsBlk)
	vote := hsBlk.(*hsBlock).block.ProposerVote()
	pm.hotstuff.OnReceiveVote(newHsVote(vote, pm.state))
	pm.hotstuff.Update(hsBlk)
}

func (pm *pacemaker) logProposal(blk hotstuff.Block) {
	qcRef := blk.Justify().Block()
	qcRefHeight := uint64(0)
	if qcRef != nil {
		qcRefHeight = qcRef.Height()
	}
	logger.I().Debugw("proposed block", "height", blk.Height(), "qcRef", qcRefHeight)
}

func (pm *pacemaker) viewChangeLoop() {
	subQC := pm.hotstuff.SubscribeNewQCHigh()
	defer subQC.Unsubscribe()

	vtimer := time.NewTimer(pm.config.ViewWidth)

	for {
		select {
		case <-pm.stopCh:
			vtimer.Stop()
			return

		case <-vtimer.C:
			pm.changeView()

		case <-time.After(pm.config.LeaderTimeout):
			logger.I().Warnw("leader timeout", "leader", pm.state.getLeaderIndex())
			pm.changeView()
			vtimer.Stop()

		case e := <-subQC.Events():
			qc := e.(hotstuff.QC)
			if ok := pm.needViewTimerResetForNewQC(qc); ok {
				vtimer.Reset(pm.config.ViewWidth)
				pm.setViewStart()
				logger.I().Infow("view timer reset", "leader", pm.state.getLeaderIndex())
			}
		}
	}
}

func (pm *pacemaker) changeView() {
	leaderIdx := pm.state.getLeaderIndex() + 1
	if leaderIdx >= pm.resources.VldStore.ValidatorCount() {
		leaderIdx = 0
	}
	pm.state.setLeaderIndex(leaderIdx)
	pm.setPendingViewChange(true)
	pm.setViewStart()
	leader := pm.resources.VldStore.GetValidator(pm.state.getLeaderIndex())
	pm.resources.MsgSvc.SendNewView(leader, pm.hotstuff.GetQCHigh().(*hsQC).qc)
	logger.I().Infow("view changed", "leader", leaderIdx, "bexec", pm.hotstuff.GetBExec().Height())
}

func (pm *pacemaker) needViewTimerResetForNewQC(qc hotstuff.QC) bool {
	proposer := qc.Block().(*hsBlock).block.Proposer()
	pidx, _ := pm.resources.VldStore.GetValidatorIndex(proposer)

	logger.I().Debugw("updated qc", "proposer", pidx, "qcRef", qc.Block().Height())

	leaderIdx := pm.state.getLeaderIndex()
	pending := pm.getPendingViewChange()
	if !pending && pidx != leaderIdx {
		pm.state.setLeaderIndex(pidx)
		return true
	}
	if pending && pidx == leaderIdx {
		pm.setPendingViewChange(false)
		return true
	}
	return false
}

func (pm *pacemaker) setViewStart() {
	pm.mtxVS.Lock()
	defer pm.mtxVS.Unlock()
	pm.viewStart = time.Now().Unix()
}

func (pm *pacemaker) getViewStart() int64 {
	pm.mtxVS.RLock()
	defer pm.mtxVS.RUnlock()
	return pm.viewStart
}

func (pm *pacemaker) setPendingViewChange(val bool) {
	pm.mtxPVC.Lock()
	defer pm.mtxPVC.Unlock()
	pm.pendingViewChange = val
}

func (pm *pacemaker) getPendingViewChange() bool {
	pm.mtxPVC.RLock()
	defer pm.mtxPVC.RUnlock()
	return pm.pendingViewChange
}
