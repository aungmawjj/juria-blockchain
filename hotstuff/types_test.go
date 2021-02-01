// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package hotstuff

import (
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

func (m *MockVote) ReplicaID() string {
	args := m.Called()
	return args.String(0)
}
