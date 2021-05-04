package util

import (
	"testing"

	"gotest.tools/assert"
)

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
			got := MajorityCount(tt.args.validatorCount)
			assert.Equal(t, tt.want, got)
		})
	}
}
