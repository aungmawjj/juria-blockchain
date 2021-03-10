// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package hotstuff

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
)

func castBlock(val interface{}) Block {
	if b, ok := val.(Block); ok {
		return b
	}
	return nil
}

type MockBlock struct {
	mock.Mock
}

func (m *MockBlock) Proposer() string {
	args := m.Called()
	return string(args.String(0))
}

func (m *MockBlock) Height() uint64 {
	args := m.Called()
	return uint64(args.Int(0))
}

func (m *MockBlock) Parent() Block {
	args := m.Called()
	return castBlock(args.Get(0))
}

func (m *MockBlock) Equal(blk Block) bool {
	args := m.Called(blk)
	return args.Bool(0)
}

func (m *MockBlock) Justify() QC {
	args := m.Called()
	return args.Get(0).(QC)
}

type MockQC struct {
	mock.Mock
}

func (m *MockQC) Block() Block {
	args := m.Called()
	return castBlock(args.Get(0))
}

type MockVote struct {
	mock.Mock
}

func (m *MockVote) Block() Block {
	args := m.Called()
	return castBlock(args.Get(0))
}

func (m *MockVote) Replica() string {
	args := m.Called()
	return args.String(0)
}

func newMockBlock(height int, parent Block, qc QC) *MockBlock {
	b := new(MockBlock)
	b.On("Height").Return(height)
	b.On("Parent").Return(parent)
	b.On("Justify").Return(qc)
	b.On("Equal", b).Return(true)
	b.On("Equal", mock.Anything).Return(false)
	return b
}

func newMockQC(blk Block) *MockQC {
	qc := new(MockQC)
	qc.On("Block").Return(blk)
	return qc
}

func newMockVote(blk Block, replica string) *MockVote {
	vote := new(MockVote)
	vote.On("Block").Return(blk)
	vote.On("Replica").Return(replica)
	return vote
}

func TestCmpBlockHeight(t *testing.T) {
	type args struct {
		b1 Block
		b2 Block
	}

	bh4 := newMockBlock(4, nil, nil)
	bh5 := newMockBlock(5, nil, nil)

	tests := []struct {
		name string
		args args
		want bool
	}{
		{"nil blocks", args{nil, nil}, false},
		{"b1 is nil", args{nil, new(MockBlock)}, false},
		{"b2 is nil", args{new(MockBlock), nil}, true},
		{"b1 is higher", args{bh5, bh4}, true},
		{"b2 is higher", args{bh4, bh5}, false},
		{"same height", args{bh4, bh4}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CmpBlockHeight(tt.args.b1, tt.args.b2); got != tt.want {
				t.Errorf("CmpBlockHeight() = %v, want %v", got, tt.want)
			}
		})
	}
}

type MockDriver struct {
	mock.Mock
}

func (m *MockDriver) CreateLeaf(ctx context.Context, parent Block, qc QC, height uint64) Block {
	args := m.Called(ctx, parent, qc, height)
	return castBlock(args.Get(0))
}

func (m *MockDriver) CreateQC(votes []Vote) QC {
	args := m.Called(votes)
	return args.Get(0).(QC)
}

func (m *MockDriver) SendProposal(blk Block) {
	m.Called(blk)
}

func (m *MockDriver) VoteBlock(blk Block) Vote {
	args := m.Called(blk)
	return args.Get(0).(Vote)
}

func (m *MockDriver) SendVote(v Vote) {
	m.Called(v)
}

func (m *MockDriver) SendNewView(qc QC) {
	m.Called(qc)
}

func (m *MockDriver) Execute(blk Block) {
	m.Called(blk)
}

func (m *MockDriver) MajorityCount() int {
	args := m.Called()
	return args.Int(0)
}
