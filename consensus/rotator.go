// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package consensus

import (
	"sync"
	"time"

	"github.com/aungmawjj/juria-blockchain/hotstuff"
	"github.com/aungmawjj/juria-blockchain/logger"
)

type rotator struct {
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

func (rot *rotator) start() {
	if rot.stopCh != nil {
		return
	}
	rot.stopCh = make(chan struct{})
	rot.setViewStart()
	rot.leaderTimeout = (rot.config.TxWaitTime + rot.config.BeatTimeout) * 5
	go rot.run()
	logger.I().Info("started rotator")
}

func (rot *rotator) stop() {
	if rot.stopCh == nil {
		return // not started yet
	}
	select {
	case <-rot.stopCh: // already stopped
		return
	default:
	}
	close(rot.stopCh)
	logger.I().Info("stopped rotator")
	rot.stopCh = nil
}

func (rot *rotator) run() {
	subQC := rot.hotstuff.SubscribeNewQCHigh()
	defer subQC.Unsubscribe()

	rot.viewTimer = time.NewTimer(rot.config.ViewWidth)
	defer rot.viewTimer.Stop()

	rot.leaderTimer = time.NewTimer(rot.leaderTimeout)
	defer rot.leaderTimer.Stop()

	for {
		select {
		case <-rot.stopCh:
			return

		case <-rot.viewTimer.C:
			rot.onViewTimeout()

		case <-rot.leaderTimer.C:
			rot.onLeaderTimeout()

		case e := <-subQC.Events():
			rot.onNewQCHigh(e.(hotstuff.QC))
		}
	}
}

func (rot *rotator) drainResetTimer(timer *time.Timer, d time.Duration) {
	rot.drainStopTimer(timer)
	timer.Reset(d)
}

func (rot *rotator) drainStopTimer(timer *time.Timer) {
	if !timer.Stop() { // timer triggered before another stop/reset call
		t := time.NewTimer(5 * time.Millisecond)
		defer t.Stop()
		select {
		case <-timer.C:
		case <-t.C: // to make sure it's not stuck more than 5ms
		}
	}
}

func (rot *rotator) onLeaderTimeout() {
	logger.I().Warnw("leader timeout", "leader", rot.state.getLeaderIndex())
	rot.leaderTimeoutCount++
	rot.changeView()
	rot.drainStopTimer(rot.viewTimer)
	if rot.leaderTimeoutCount > rot.state.getFaultyCount() {
		rot.leaderTimer.Stop()
		rot.setPendingViewChange(false)
	} else {
		rot.leaderTimer.Reset(rot.leaderTimeout)
	}
}

func (rot *rotator) onViewTimeout() {
	rot.changeView()
	rot.drainResetTimer(rot.leaderTimer, rot.leaderTimeout)
}

func (rot *rotator) changeView() {
	leaderIdx := rot.nextLeader()
	rot.state.setLeaderIndex(leaderIdx)
	rot.setPendingViewChange(true)
	rot.setViewStart()
	leader := rot.resources.VldStore.GetValidator(rot.state.getLeaderIndex())
	rot.resources.MsgSvc.SendNewView(leader, rot.hotstuff.GetQCHigh().(*hsQC).qc)
	logger.I().Infow("view changed",
		"leader", leaderIdx, "qc", qcRefHeight(rot.hotstuff.GetQCHigh()))
}

func (rot *rotator) nextLeader() int {
	leaderIdx := rot.state.getLeaderIndex() + 1
	if leaderIdx >= rot.resources.VldStore.ValidatorCount() {
		leaderIdx = 0
	}
	return leaderIdx
}

func (rot *rotator) onNewQCHigh(qc hotstuff.QC) {
	rot.state.setQC(qc.(*hsQC).qc)
	proposer := rot.resources.VldStore.GetValidatorIndex(qcRefProposer(qc))
	logger.I().Debugw("updated qc", "proposer", proposer, "qc", qcRefHeight(qc))
	var ltreset, vtreset bool
	if proposer == rot.state.getLeaderIndex() { // if qc is from current leader
		ltreset = true
	}
	if rot.isNewViewApproval(proposer) {
		ltreset = true
		vtreset = true
		rot.approveViewLeader(proposer)
	}
	if ltreset {
		rot.drainResetTimer(rot.leaderTimer, rot.leaderTimeout)
	}
	if vtreset {
		rot.drainResetTimer(rot.viewTimer, rot.config.ViewWidth)
	}
}

func (rot *rotator) isNewViewApproval(proposer int) bool {
	leaderIdx := rot.state.getLeaderIndex()
	pending := rot.getPendingViewChange()
	return (!pending && proposer != leaderIdx) || // node first run or out of sync
		(pending && proposer == leaderIdx) // expecting leader
}

func (rot *rotator) approveViewLeader(proposer int) {
	rot.setPendingViewChange(false)
	rot.state.setLeaderIndex(proposer)
	rot.setViewStart()
	logger.I().Infow("approved leader", "leader", rot.state.getLeaderIndex())
	rot.leaderTimeoutCount = 0
}

func (rot *rotator) setViewStart() {
	rot.mtxVS.Lock()
	defer rot.mtxVS.Unlock()
	rot.viewStart = time.Now().Unix()
}

func (rot *rotator) getViewStart() int64 {
	rot.mtxVS.RLock()
	defer rot.mtxVS.RUnlock()
	return rot.viewStart
}

func (rot *rotator) setPendingViewChange(val bool) {
	rot.mtxPVC.Lock()
	defer rot.mtxPVC.Unlock()
	rot.pendingViewChange = val
}

func (rot *rotator) getPendingViewChange() bool {
	rot.mtxPVC.RLock()
	defer rot.mtxPVC.RUnlock()
	return rot.pendingViewChange
}
