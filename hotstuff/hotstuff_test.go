// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package hotstuff

import (
	"context"
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
			hs.state.init(tt.fields.bLeaf, tt.fields.qcHigh)

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
	hs := new(Hotstuff)
	hs.driver = driver
	hs.state.init(b0, q0)

	b1 := newMockBlock(11, b0, q0)

	driver.On("CreateLeaf", mock.Anything, b0, q0, b0.Height()+1).Once().Return(b1)
	driver.On("BroadcastProposal", b1).Once()

	hs.OnPropose(context.Background())

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
	hs := new(Hotstuff)
	hs.driver = driver
	hs.state.init(b0, q0)

	driver.On("CreateLeaf", mock.Anything, b0, q0, b0.Height()+1).Once().Return(nil)

	hs.OnPropose(context.Background())

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
	hs := new(Hotstuff)
	hs.driver = driver
	hs.state.init(b0, q0)
	driver.On("CreateLeaf", mock.Anything, b0, q0, b0.Height()+1).Once().Return(b1)
	driver.On("BroadcastProposal", b1).Once()
	hs.OnPropose(context.Background())
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

	tests := []struct {
		name    string
		vHeight uint64
		bLock   Block
		bNew    Block
		want    bool
	}{
		{"proposal 1", 10, b0, b1, true},
		{"proposal 2", 11, b0, b2, true},
		{"proposal 3", 12, b0, b3, true},
		{"proposal 4", 13, b1, b4, true},
		{"proposal not higher", 13, b1, b3, false},
		{"trigger liveness rule", 13, b1, bf4, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hs := new(Hotstuff)
			hs.setVHeight(tt.vHeight)
			hs.setBLock(tt.bLock)

			assert.Equal(t, tt.want, hs.CanVote(tt.bNew))
		})
	}
}
