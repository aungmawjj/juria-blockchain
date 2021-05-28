// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package merkle

import (
	"crypto"
	"crypto/sha1"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPosition(t *testing.T) {
	tests := []struct {
		name  string
		level uint8
		index *big.Int
		want  []byte
	}{
		{"level 0, index 0", 0, big.NewInt(0), []byte{0, 0}},
		{"index 0", 1, big.NewInt(0), []byte{1, 0}},
		{"index max 8 bit", 1, big.NewInt(255), []byte{1, 255}},
		{"index first 16 bit", 1, big.NewInt(256), []byte{1, 1, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			p := NewPosition(tt.level, tt.index)

			assert.EqualValues(tt.want, p.Bytes())
			assert.Equal(string(tt.want), p.String())

			p1 := UnmarshalPosition(p.Bytes())

			assert.Equal(p.Level(), p1.Level())
			assert.Equal(0, p.Index().Cmp(p1.Index()))
		})
	}
}

func TestGroup_Load_Sum(t *testing.T) {
	s1 := NewMapStore()
	s2 := storeWith3Nodes()

	tests := []struct {
		name           string
		bfactor        uint8
		store          Store
		parentPosition *Position
		empty          bool
	}{
		{"empty store", 2, s1, NewPosition(1, big.NewInt(0)), true},
		{"2 leaves with bf 2", 2, s2, NewPosition(1, big.NewInt(0)), false},
		{"2 leaves with bf 4", 4, s2, NewPosition(1, big.NewInt(0)), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			b := NewGroup(crypto.SHA1, NewTreeCalc(tt.bfactor), tt.store, tt.parentPosition)
			b.Load(big.NewInt(100))

			assert.EqualValues(tt.bfactor, len(b.nodes))
			assert.Equal(tt.empty, b.IsEmpty())

			assert.Equal(tt.store.GetNode(tt.parentPosition), b.MakeParent().Data)
		})
	}
}

func TestGroup_SetNode(t *testing.T) {
	store := storeWith3Nodes()

	b := NewGroup(crypto.SHA1, NewTreeCalc(2), store, NewPosition(1, big.NewInt(0)))
	b.Load(big.NewInt(2))
	b.SetNode(&Node{
		Position: NewPosition(0, big.NewInt(1)),
		Data:     []byte{3, 3},
	})

	assert := assert.New(t)
	assert.False(b.IsEmpty())
	assert.Equal(sha1Sum([]byte{1, 1, 3, 3}), b.MakeParent().Data)
}

//   h(1,1,2,2)
//    /      \
// [1,1]    [2,2]
func storeWith3Nodes() Store {
	s2 := NewMapStore()

	upd := &UpdateResult{
		LeafCount: big.NewInt(2),
		Height:    2,
		Leaves: []*Node{
			{NewPosition(0, big.NewInt(0)), []byte{1, 1}},
			{NewPosition(0, big.NewInt(1)), []byte{2, 2}},
		},
		Branches: []*Node{
			{NewPosition(1, big.NewInt(0)), sha1Sum([]byte{1, 1, 2, 2})},
		},
	}

	s2.CommitUpdate(upd)
	return s2
}

func sha1Sum(b []byte) []byte {
	h := sha1.New()
	h.Write(b)
	return h.Sum(nil)
}
