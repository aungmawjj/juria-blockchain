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
		delayT := time.NewTimer(d)
		select {
		case <-pm.stopCh:
			return
		case <-delayT.C:
		case <-subQC.Events():
		}
		delayT.Stop()
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

	leaderTimeout := (pm.config.TxWaitTime + pm.config.BeatDelay) * 5

	viewTimer := time.NewTimer(pm.config.ViewWidth)
	defer viewTimer.Stop()
	leaderTimer := time.NewTimer(leaderTimeout)
	defer leaderTimer.Stop()

	for {
		select {
		case <-pm.stopCh:
			return

		case <-viewTimer.C:
			if !pm.getPendingViewChange() {
				pm.changeView()
				drainReset(leaderTimer, leaderTimeout)
			}

		case <-leaderTimer.C:
			pm.onLeaderTimeout()
			viewTimer.Stop()
			leaderTimer.Reset(leaderTimeout)

		case e := <-subQC.Events():
			ltreset, vtreset := pm.onNewQCHigh(e.(hotstuff.QC))
			if ltreset {
				drainReset(leaderTimer, leaderTimeout)
			}
			if vtreset {
				viewTimer.Reset(pm.config.ViewWidth)
			}
		}
	}
}

func drainReset(timer *time.Timer, d time.Duration) {
	if !timer.Stop() {
		<-timer.C
	}
	timer.Reset(d)
}

func (pm *pacemaker) onLeaderTimeout() {
	logger.I().Warnw("leader timeout", "leader", pm.state.getLeaderIndex())
	pm.changeView()
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
	logger.I().Infow("view changed",
		"leader", leaderIdx, "qc", qcRefHeight(pm.hotstuff.GetQCHigh()))
}

func (pm *pacemaker) onNewQCHigh(qc hotstuff.QC) (leaderTReset, viewTReset bool) {
	pm.state.setQC(qc.(*hsQC).qc)
	pidx := pm.resources.VldStore.GetValidatorIndex(qc.Block().(*hsBlock).block.Proposer())
	logger.I().Debugw("updated qc", "proposer", pidx, "qc", qcRefHeight(qc))
	if pidx == pm.state.getLeaderIndex() { // if qc is from current leader
		leaderTReset = true
	}
	if pm.isFirstQCForCurrentView(pidx) {
		viewTReset = true
		pm.approveViewLeader(pidx)
	}
	return leaderTReset, viewTReset
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
