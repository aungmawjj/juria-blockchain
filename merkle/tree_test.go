// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package merkle

import (
	"crypto"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTree(t *testing.T) {
	tests := []struct {
		name string
		opts TreeOptions
		want uint8
	}{
		{"branch factor < 2", TreeOptions{1, crypto.SHA1}, 2},
		{"branch factor >= 2", TreeOptions{4, crypto.SHA1}, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := NewTree(nil, tt.opts)
			assert.Equal(t, tt.want, tree.bfactor)
		})
	}
}
