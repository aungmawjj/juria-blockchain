// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package hotstuff

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHotstuff_UpdateQCHigh(t *testing.T) {
	q0 := newMockQC(nil)
	b0 := newMockBlock(10, nil, q0)

	b1 := newMockBlock(11, b0, q0)
	q1 := newMockQC(b1)

	b2 := newMockBlock(12, b1, q1)
	q2 := newMockQC(b2)

	type args struct {
		qc QC
	}
	type fields struct {
		bLeaf  Block
		qcHigh QC
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   fields
	}{
		{
			"same as qcHigh",
			fields{b0, q0},
			args{q0},
			fields{b0, q0},
		},
		{
			"higher qc1",
			fields{b0, q0},
			args{q1},
			fields{b1, q1},
		},
		{
			"higher qc2",
			fields{b1, q1},
			args{q2},
			fields{b2, q2},
		},
		{
			"lower qc",
			fields{b1, q1},
			args{q0},
			fields{b1, q1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hs := new(Hotstuff)
			hs.state = newState(tt.fields.bLeaf, tt.fields.qcHigh)

			hs.UpdateQCHigh(tt.args.qc)

			assert := assert.New(t)
			assert.Equal(tt.want.bLeaf, hs.GetBLeaf())
			assert.Equal(tt.want.qcHigh, hs.GetQCHigh())
		})
	}
}

func TestHotstuff_SuccessfulPropose(t *testing.T) {
	q0 := newMockQC(nil)
	b0 := newMockBlock(10, nil, q0)

	driver := new(MockDriver)
	hs := New(driver, b0, q0)

	b1 := newMockBlock(11, b0, q0)

	driver.On("CreateLeaf", b0, q0, b0.Height()+1).Once().Return(b1)
	driver.On("BroadcastProposal", b1).Once()

	hs.OnPropose()

	driver.AssertExpectations(t)

	assert := assert.New(t)
	assert.Equal(b1, hs.GetBLeaf())
	assert.True(hs.IsProposing())

	driver.On("MajorityCount").Return(3)

	v1 := newMockVote(b1, "r1")
	hs.OnReceiveVote(v1)

	assert.Equal([]Vote{v1}, hs.GetVotes())
}

func TestHotstuff_FailedPropose(t *testing.T) {
	q0 := newMockQC(nil)
	b0 := newMockBlock(10, nil, q0)

	driver := new(MockDriver)
	hs := New(driver, b0, q0)

	driver.On("CreateLeaf", b0, q0, b0.Height()+1).Once().Return(nil)

	hs.OnPropose()

	driver.AssertExpectations(t)
	driver.AssertNotCalled(t, "BroadcastProposal")

	assert := assert.New(t)
	assert.Equal(b0, hs.GetBLeaf())
	assert.False(hs.IsProposing())
}

func TestHotstuff_OnReceiveVote(t *testing.T) {
	q0 := newMockQC(nil)
	b0 := newMockBlock(10, nil, q0)
	b1 := newMockBlock(11, b0, q0)
	q1 := newMockQC(b1)

	assert := assert.New(t)

	driver := new(MockDriver)
	hs := New(driver, b0, q0)

	driver.On("CreateLeaf", b0, q0, b0.Height()+1).Once().Return(b1)
	driver.On("BroadcastProposal", b1).Once()
	hs.OnPropose()
	driver.On("MajorityCount").Return(2)

	v1 := newMockVote(b1, "r1")
	hs.OnReceiveVote(v1)

	driver.AssertNotCalled(t, "CreateQC")
	assert.Equal(1, hs.GetVoteCount())

	v1Dup := newMockVote(b1, "r1")
	hs.OnReceiveVote(v1Dup)

	driver.AssertNotCalled(t, "CreateQC")
	assert.Equal(1, hs.GetVoteCount())

	driver.On("CreateQC", mock.Anything).Return(q1)

	v2 := newMockVote(b1, "r2")
	hs.OnReceiveVote(v2)

	driver.AssertExpectations(t)
	assert.Equal(0, hs.GetVoteCount())
	assert.False(hs.IsProposing())
	assert.Equal(q1, hs.GetQCHigh())
}

func TestHotstuff_CanVote(t *testing.T) {
	q0 := newMockQC(nil)
	b0 := newMockBlock(10, nil, q0) // bLock

	b1 := newMockBlock(11, b0, q0)
	q1 := newMockQC(b1)

	b2 := newMockBlock(12, b1, q1)
	q2 := newMockQC(b2)

	b3 := newMockBlock(13, b2, q2)
	q3 := newMockQC(b3)

	b4 := newMockBlock(14, b3, q3)

	bf1 := newMockBlock(11, b0, q0)
	qf1 := newMockQC(bf1)

	bf2 := newMockBlock(12, bf1, qf1)
	qf2 := newMockQC(bf2)

	bf3 := newMockBlock(13, bf1, qf2)
	bf4 := newMockBlock(14, bf3, qf2)
	_ = bf4

	tests := []struct {
		name  string
		bVote Block
		bLock Block
		bNew  Block
		want  bool
	}{
		{"proposal 1", b0, b0, b1, true},
		{"proposal 2", b1, b0, b2, true},
		{"proposal 3", b2, b0, b3, true},
		{"proposal 4", b3, b1, b4, true},
		{"proposal not higher", b3, b1, b3, false},
		{"trigger liveness rule", b3, b1, bf4, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hs := new(Hotstuff)
			hs.state = new(state)
			hs.setBVote(tt.bVote)
			hs.setBLock(tt.bLock)

			assert.Equal(t, tt.want, hs.CanVote(tt.bNew))
		})
	}
}

func TestHotstuff_Update(t *testing.T) {
	q0 := newMockQC(nil)
	b0 := newMockBlock(10, nil, q0) // bLock

	b1 := newMockBlock(11, b0, q0)
	q1 := newMockQC(b1)

	b2 := newMockBlock(12, b1, q1)
	q2 := newMockQC(b2)

	b3 := newMockBlock(13, b2, q2)
	q3 := newMockQC(b3)

	b4 := newMockBlock(14, b3, q3)
	_ = b4

	bf0 := newMockBlock(10, nil, q0)

	bb1 := newMockBlock(11, b0, q0)
	qq1 := newMockQC(bb1)

	bb2 := newMockBlock(12, bb1, qq1) // could not get qc for bb2

	bb3 := newMockBlock(13, bb2, qq1)
	qq3 := newMockQC(bb3)

	bb4 := newMockBlock(14, bb3, qq3)
	qq4 := newMockQC(bb4)

	bb5 := newMockBlock(15, bb4, qq4)
	qq5 := newMockQC(bb5)

	bb6 := newMockBlock(16, bb5, qq5)
	_ = bb6

	hs0 := new(Hotstuff)
	hs0.state = newState(b0, q0)

	hs1 := new(Hotstuff)
	hs1.state = newState(b0, q1)

	hs2 := new(Hotstuff)
	hs2.state = newState(b0, q2)
	hs2.setBLock(b1)

	hs3 := new(Hotstuff)
	hs3.state = newState(b2, q2)

	tests := []struct {
		name      string
		hs        *Hotstuff
		bNew      Block
		execCount int
		qcHigh    QC
		bLock     Block
		bExec     Block
	}{
		{"proposal 1", hs0, b1, 0, q0, b0, b0},
		{"proposal dup", hs0, bf0, 0, q0, b0, b0},
		{"proposal 2", hs0, b2, 0, q1, b0, b0},
		{"proposal 3", hs1, b3, 0, q2, b1, b0},
		{"proposal 4", hs2, b4, 1, q3, b2, b1},
		{"not three chain", hs0, bb5, 0, qq4, bb3, b0},
		{"exec debts", hs0, bb6, 3, qq5, bb4, bb3},
		{"three chain but invalid commit phase", hs3, b4, 0, q3, b2, b2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver := new(MockDriver)
			tt.hs.driver = driver
			if tt.execCount > 0 {
				driver.On("Commit", mock.Anything).Times(tt.execCount)
			}
			tt.hs.Update(tt.bNew)

			assert := assert.New(t)
			driver.AssertExpectations(t)
			assert.Equal(tt.qcHigh, tt.hs.GetQCHigh())
			assert.Equal(tt.bLock, tt.hs.GetBLock())
			assert.Equal(tt.bExec, tt.hs.GetBExec())
		})
	}
}
