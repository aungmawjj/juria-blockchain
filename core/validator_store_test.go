// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package core

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"gotest.tools/assert"
)

type MockValidatorStore struct {
	mock.Mock
}

var _ ValidatorStore = (*MockValidatorStore)(nil)

func (m *MockValidatorStore) ValidatorCount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockValidatorStore) MajorityCount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockValidatorStore) IsValidator(pubKey *PublicKey) bool {
	args := m.Called(pubKey)
	return args.Bool(0)
}

func (m *MockValidatorStore) GetValidator(idx int) *PublicKey {
	args := m.Called(idx)
	val := args.Get(0)
	if val == nil {
		return nil
	}
	return val.(*PublicKey)
}

func (m *MockValidatorStore) GetValidatorIndex(pubKey *PublicKey) int {
	args := m.Called(pubKey)
	return args.Int(0)
}

func TestMajorityCount(t *testing.T) {
	type args struct {
		validatorCount int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"single node", args{1}, 1},
		{"exact factor", args{4}, 3},  // n = 3f+1, f=1
		{"exact factor", args{10}, 7}, // f=3, m=10-3
		{"middle", args{12}, 9},       // f=3, m=12-3
		{"middle", args{14}, 10},      // f=4, m=14-4
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := majorityCount(tt.args.validatorCount)
			assert.Equal(t, tt.want, got)
		})
	}
}
