// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package hotstuff

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type state struct {
	vHeight atomic.Value
	bLock   atomic.Value
	bExec   atomic.Value
	qcHigh  atomic.Value
	bLeaf   atomic.Value

	proposal Block
	votes    map[string]Vote
	pMtx     sync.RWMutex
}

func (s *state) init(b0 Block, q0 QC) {
	s.setVHeight(b0.Height())
	s.setBLock(b0)
	s.setBExec(b0)
	s.setBLeaf(b0)
	s.setQCHigh(q0)
}

func (s *state) setVHeight(height uint64) { s.vHeight.Store(height) }
func (s *state) setBLock(b Block)         { s.bLock.Store(b) }
func (s *state) setBExec(b Block)         { s.bExec.Store(b) }
func (s *state) setBLeaf(bNew Block)      { s.bLeaf.Store(bNew) }
func (s *state) setQCHigh(qcHigh QC)      { s.qcHigh.Store(qcHigh) }

func (s *state) GetVHeight() uint64 {
	return s.vHeight.Load().(uint64)
}

func (s *state) GetBLock() Block {
	return s.bLock.Load().(Block)
}

func (s *state) GetBExec() Block {
	return s.bExec.Load().(Block)
}

func (s *state) GetBLeaf() Block {
	return s.bLeaf.Load().(Block)
}

func (s *state) GetQCHigh() QC {
	return s.qcHigh.Load().(QC)
}

func (s *state) IsProposing() bool {
	s.pMtx.RLock()
	defer s.pMtx.RUnlock()

	return s.proposal != nil
}

func (s *state) startProposal(b Block) {
	s.pMtx.Lock()
	defer s.pMtx.Unlock()

	s.proposal = b
	s.votes = make(map[string]Vote)
}

func (s *state) stopProposal(b Block) {
	s.pMtx.Lock()
	defer s.pMtx.Unlock()

	s.proposal = nil
	s.votes = nil
}

func (s *state) addVote(v Vote) error {
	s.pMtx.Lock()
	defer s.pMtx.Unlock()

	if s.proposal == nil {
		return fmt.Errorf("no proposal in progress")
	}
	if !s.proposal.Equal(v.Block()) {
		return fmt.Errorf("not same block")
	}
	key := v.Replica()
	if _, found := s.votes[key]; found {
		return fmt.Errorf("duplicate vote")
	}
	s.votes[key] = v
	return nil
}

func (s *state) GetVotes() []Vote {
	s.pMtx.RLock()
	defer s.pMtx.RUnlock()

	votes := make([]Vote, 0, len(s.votes))
	for _, v := range s.votes {
		votes = append(votes, v)
	}
	return votes
}
