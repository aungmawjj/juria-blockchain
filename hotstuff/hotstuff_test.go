// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package hotstuff

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHotstuff_OnReceiveNewView(t *testing.T) {
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

			hs.updateQCHigh(tt.args.qc)

			assert := assert.New(t)
			assert.Equal(tt.want.bLeaf, hs.GetBLeaf())
			assert.Equal(tt.want.qcHigh, hs.GetQCHigh())
		})
	}
}

func TestHotstuff_OnNextSyncView(t *testing.T) {

	q0 := newMockQC(nil)
	b0 := newMockBlock(10, nil, q0)

	b1 := newMockBlock(11, b0, q0)
	q1 := newMockQC(b1)

	driver := new(MockDriver)

	hs := new(Hotstuff)
	hs.driver = driver
	hs.state.init(b0, q0)

	driver.On("SendNewView", q0).Once()
	hs.OnNextSyncView()

	driver.AssertExpectations(t)

	hs.state.init(b1, q1)

	driver.On("SendNewView", q1).Once()
	hs.OnNextSyncView()

	driver.AssertExpectations(t)
}

func TestHotstuff_OnPropose(t *testing.T) {
	q0 := newMockQC(nil)
	b0 := newMockBlock(10, nil, q0)

	b1 := newMockBlock(11, b0, q0)

	driver := new(MockDriver)

	hs := new(Hotstuff)
	hs.driver = driver
	hs.state.init(b0, q0)

	driver.On("CreateLeaf", mock.Anything, b0, q0, b0.Height()+1).Once().Return(b1)
	driver.On("SendProposal", b1).Once()

	hs.OnPropose(context.Background())

	driver.AssertExpectations(t)

	assert := assert.New(t)
	assert.Equal(b1, hs.GetBLeaf())
	assert.True(hs.IsProposing())

	hs = new(Hotstuff)
	hs.driver = driver
	hs.state.init(b0, q0)

	driver.On("CreateLeaf", mock.Anything, b0, q0, b0.Height()+1).Once().Return(nil)

	hs.OnPropose(context.Background())

	driver.AssertExpectations(t)
	driver.AssertNotCalled(t, "SendProposal")

	assert.Equal(b0, hs.GetBLeaf())
	assert.False(hs.IsProposing())
}
