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

	leaderTimeout time.Duration
	leaderTimer   *time.Timer
	viewTimer     *time.Timer

	// start timestamp in second of current view
	viewStart int64
	mtxVS     sync.RWMutex

	// true when view changed before the next leader is approved
	pendingViewChange bool
	mtxPVC            sync.RWMutex

	leaderTimeoutCount int

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

	for {
		pm.onBeat()
		d := pm.config.BeatDelay
		if pm.resources.TxPool.GetStatus().Total == 0 {
			d += pm.config.TxWaitTime
		}
		delayT := time.NewTimer(d)
		subQC := pm.hotstuff.SubscribeNewQCHigh()

		select {
		case <-pm.stopCh:
			return
		case <-delayT.C:
		case <-subQC.Events():
		}
		delayT.Stop()
		subQC.Unsubscribe()
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

	pm.leaderTimeout = (pm.config.TxWaitTime + pm.config.BeatDelay) * 5

	pm.viewTimer = time.NewTimer(pm.config.ViewWidth)
	defer pm.viewTimer.Stop()

	pm.leaderTimer = time.NewTimer(pm.leaderTimeout)
	defer pm.leaderTimer.Stop()

	for {
		select {
		case <-pm.stopCh:
			return

		case <-pm.viewTimer.C:
			pm.onViewTimeout()

		case <-pm.leaderTimer.C:
			pm.onLeaderTimeout()

		case e := <-subQC.Events():
			pm.onNewQCHigh(e.(hotstuff.QC))
		}
	}
}

func (pm *pacemaker) drainResetTimer(timer *time.Timer, d time.Duration) {
	pm.drainStopTimer(timer)
	timer.Reset(d)
}

func (pm *pacemaker) drainStopTimer(timer *time.Timer) {
	if !timer.Stop() { // timer triggered before another stop/reset call
		t := time.NewTimer(5 * time.Millisecond)
		defer t.Stop()
		select {
		case <-timer.C:
		case <-t.C: // to make sure it's not stuck more than 5ms
		}
	}
}

func (pm *pacemaker) onLeaderTimeout() {
	logger.I().Warnw("leader timeout", "leader", pm.state.getLeaderIndex())
	pm.leaderTimeoutCount++
	pm.changeView()
	pm.drainStopTimer(pm.viewTimer)
	if pm.leaderTimeoutCount > pm.state.getFaultyCount() {
		pm.leaderTimer.Stop()
		pm.setPendingViewChange(false)
	} else {
		pm.leaderTimer.Reset(pm.leaderTimeout)
	}
}

func (pm *pacemaker) onViewTimeout() {
	pm.changeView()
	pm.drainResetTimer(pm.leaderTimer, pm.leaderTimeout)
}

func (pm *pacemaker) changeView() {
	leaderIdx := pm.nextLeader()
	pm.state.setLeaderIndex(leaderIdx)
	pm.setPendingViewChange(true)
	pm.setViewStart()
	leader := pm.resources.VldStore.GetValidator(pm.state.getLeaderIndex())
	pm.resources.MsgSvc.SendNewView(leader, pm.hotstuff.GetQCHigh().(*hsQC).qc)
	logger.I().Infow("view changed",
		"leader", leaderIdx, "qc", qcRefHeight(pm.hotstuff.GetQCHigh()))
}

func (pm *pacemaker) nextLeader() int {
	leaderIdx := pm.state.getLeaderIndex() + 1
	if leaderIdx >= pm.resources.VldStore.ValidatorCount() {
		leaderIdx = 0
	}
	return leaderIdx
}

func (pm *pacemaker) onNewQCHigh(qc hotstuff.QC) {
	pm.state.setQC(qc.(*hsQC).qc)
	proposer := pm.resources.VldStore.GetValidatorIndex(qcRefProposer(qc))
	logger.I().Debugw("updated qc", "proposer", proposer, "qc", qcRefHeight(qc))
	var ltreset, vtreset bool
	if proposer == pm.state.getLeaderIndex() { // if qc is from current leader
		ltreset = true
	}
	if pm.isNewViewApproval(proposer) {
		ltreset = true
		vtreset = true
		pm.approveViewLeader(proposer)
	}
	if ltreset {
		pm.drainResetTimer(pm.leaderTimer, pm.leaderTimeout)
	}
	if vtreset {
		pm.drainResetTimer(pm.viewTimer, pm.config.ViewWidth)
	}
}

func (pm *pacemaker) isNewViewApproval(proposer int) bool {
	leaderIdx := pm.state.getLeaderIndex()
	pending := pm.getPendingViewChange()
	return (!pending && proposer != leaderIdx) || // node first run or out of sync
		(pending && proposer == leaderIdx) // expecting leader
}

func (pm *pacemaker) approveViewLeader(proposer int) {
	pm.setPendingViewChange(false)
	pm.state.setLeaderIndex(proposer)
	pm.setViewStart()
	logger.I().Infow("approved leader", "leader", pm.state.getLeaderIndex())
	pm.leaderTimeoutCount = 0
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
