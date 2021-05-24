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
	state     *state
	hotstuff  *hotstuff.Hotstuff

	// leader wait for this duration after proposal to get a qc
	// leader will propose next block when a new qc is created or after beat delay
	beatDelay time.Duration

	viewWidth time.Duration

	// the validators change view if the leader failed to create next qc in this duration
	leaderTimeout time.Duration

	// pendingViewChange is set true after viewChange
	// reset view timer if pendingViewChange is true on next qc update and set to false
	pendingViewChange bool
	stopCh            chan struct{}
}

func (pm *pacemaker) start() {
	if pm.stopCh != nil {
		select {
		case <-pm.stopCh: // confirm stopping
			break
		default: // pm not stopping, cannot start
			return
		}
	}
	pm.stopCh = make(chan struct{})
	pm.pendingViewChange = false
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
}

func (pm *pacemaker) beatLoop() {
	subQC := pm.hotstuff.SubscribeNewQCHigh()
	defer subQC.Unsubscribe()

	for {
		pm.onBeat()
		select {
		case <-pm.stopCh:
			return
		case <-time.After(pm.beatDelay):
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
	logger.I().Debugw("proposed block",
		"height", hsBlk.Height(), "qcRef", hsBlk.Justify().Block().Height(),
	)

	vote := hsBlk.(*hsBlock).block.ProposerVote()
	pm.hotstuff.OnReceiveVote(newHsVote(vote, pm.state))
	pm.hotstuff.Update(hsBlk)
}

func (pm *pacemaker) viewChangeLoop() {
	subQC := pm.hotstuff.SubscribeNewQCHigh()
	defer subQC.Unsubscribe()

	vtimer := time.NewTimer(pm.viewWidth)

	for {
		select {
		case <-pm.stopCh:
			vtimer.Stop()
			return

		case <-vtimer.C:
			pm.changeView()

		case <-time.After(pm.leaderTimeout):
			logger.I().Warnw("leader timeout", "leader", pm.state.getLeaderIndex())
			pm.changeView()
			vtimer.Stop()

		case e := <-subQC.Events():
			qc := e.(hotstuff.QC)
			if ok := pm.needViewTimerResetForNewQC(qc); ok {
				vtimer.Reset(pm.viewWidth)
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
	pm.pendingViewChange = true
	leader := pm.resources.VldStore.GetValidator(pm.state.getLeaderIndex())
	pm.resources.MsgSvc.SendNewView(leader, pm.hotstuff.GetQCHigh().(*hsQC).qc)

	logger.I().Infow("view changed", "leader", leaderIdx, "bexec", pm.hotstuff.GetBExec().Height())
}

func (pm *pacemaker) needViewTimerResetForNewQC(qc hotstuff.QC) bool {
	proposer := qc.Block().(*hsBlock).block.Proposer()
	pidx, _ := pm.resources.VldStore.GetValidatorIndex(proposer)

	logger.I().Debugw("updated qc", "proposer", pidx, "qcRef", qc.Block().Height())

	leaderIdx := pm.state.getLeaderIndex()
	if !pm.pendingViewChange && pidx != leaderIdx {
		pm.state.setLeaderIndex(pidx)
		return true
	}
	if pm.pendingViewChange && pidx == leaderIdx {
		pm.pendingViewChange = false
		return true
	}
	return false
}
