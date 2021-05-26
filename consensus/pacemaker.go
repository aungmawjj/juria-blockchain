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

	viewTimer   *time.Timer
	leaderTimer *time.Timer

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
	pm.propose()
}

func (pm *pacemaker) propose() {
	blk := pm.hotstuff.OnPropose()
	logger.I().Debugw("proposed block", "height", blk.Height(), "qc", qcRefHeight(blk.Justify()))
	vote := blk.(*hsBlock).block.ProposerVote()
	pm.hotstuff.OnReceiveVote(newHsVote(vote, pm.state))
	pm.hotstuff.Update(blk)
}

func (pm *pacemaker) viewChangeLoop() {
	subQC := pm.hotstuff.SubscribeNewQCHigh()
	defer subQC.Unsubscribe()

	pm.viewTimer = time.NewTimer(pm.config.ViewWidth)
	defer pm.viewTimer.Stop()

	pm.leaderTimer = time.NewTimer(pm.config.LeaderTimeout)
	defer pm.leaderTimer.Stop()

	for {
		select {
		case <-pm.stopCh:
			return

		case <-pm.viewTimer.C:
			pm.changeView()

		case <-pm.leaderTimer.C:
			pm.onLeaderTimeout()

		case e := <-subQC.Events():
			pm.onNewQCHigh(e.(hotstuff.QC))
		}
	}
}

func (pm *pacemaker) onLeaderTimeout() {
	logger.I().Warnw("leader timeout", "leader", pm.state.getLeaderIndex())
	pm.changeView()
	pm.viewTimer.Stop()
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
	pm.leaderTimer.Reset(pm.config.LeaderTimeout)
	logger.I().Infow("view changed",
		"leader", leaderIdx, "qc", qcRefHeight(pm.hotstuff.GetQCHigh()))
}

func (pm *pacemaker) onNewQCHigh(qc hotstuff.QC) {
	pidx := pm.getProposerIndexForQC(qc)
	logger.I().Debugw("updated qc", "proposer", pidx, "qc", qcRefHeight(qc))
	if pidx == pm.state.getLeaderIndex() {
		pm.leaderTimer.Reset(pm.config.LeaderTimeout)
	}
	if pm.isFirstQCForCurrentView(pidx) {
		pm.approveViewLeader(pidx)
	}
}

func (pm *pacemaker) getProposerIndexForQC(qc hotstuff.QC) int {
	proposer := qc.Block().(*hsBlock).block.Proposer()
	pidx, _ := pm.resources.VldStore.GetValidatorIndex(proposer)
	return pidx
}

func (pm *pacemaker) isFirstQCForCurrentView(proposer int) bool {
	leaderIdx := pm.state.getLeaderIndex()
	pending := pm.getPendingViewChange()
	return (!pending && proposer != leaderIdx) || (pending && proposer == leaderIdx)
}

func (pm *pacemaker) approveViewLeader(proposer int) {
	pm.setPendingViewChange(false)
	pm.state.setLeaderIndex(proposer)
	pm.setViewStart()
	pm.viewTimer.Reset(pm.config.ViewWidth)
	logger.I().Infow("approved leader", "leader", pm.state.getLeaderIndex())
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
