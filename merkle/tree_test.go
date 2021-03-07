// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package merkle

import (
	"crypto"
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

func TestNewTree(t *testing.T) {
	tests := []struct {
		name string
		opts TreeOptions
		want *big.Int
	}{
		{"branch factor < 2", TreeOptions{1, crypto.SHA1}, big.NewInt(2)},
		{"branch factor >= 2", TreeOptions{4, crypto.SHA1}, big.NewInt(4)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := NewTree(tt.opts)
			assert.Equal(t, 0, tt.want.Cmp(tree.bfactor))
		})
	}
}
