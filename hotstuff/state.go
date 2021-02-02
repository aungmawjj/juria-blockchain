// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package hotstuff

import (
	"sync/atomic"
)

type state struct {
	vHeight atomic.Value
	bLock   atomic.Value
	bExec   atomic.Value
	qcHigh  atomic.Value
	bLeaf   atomic.Value
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

func (s *state) GetVHeight() uint64 { return s.vHeight.Load().(uint64) }
func (s *state) GetBLock() Block    { return s.bLock.Load().(Block) }
func (s *state) GetBExec() Block    { return s.bExec.Load().(Block) }
func (s *state) GetBLeaf() Block    { return s.bLeaf.Load().(Block) }
func (s *state) GetQCHigh() QC      { return s.qcHigh.Load().(QC) }
