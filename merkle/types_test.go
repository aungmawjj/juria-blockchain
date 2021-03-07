// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package merkle

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPosition_Bytes(t *testing.T) {
	tests := []struct {
		name  string
		level uint8
		index *big.Int
		want  []byte
	}{
		{"level 0, index 0", 0, big.NewInt(0), []byte{0}},
		{"index 0", 1, big.NewInt(0), []byte{1}},
		{"index max 8 bit", 1, big.NewInt(255), []byte{1, 255}},
		{"index first 16 bit", 1, big.NewInt(256), []byte{1, 1, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Position{tt.level, tt.index}
			assert.EqualValues(t, tt.want, p.Bytes())
		})
	}
}
