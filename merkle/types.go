// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package merkle

import "math/big"

// Position of a node in the tree
type Position struct {
	level uint8
	index *big.Int
	bytes []byte
	str   string
}

// UnmarshalPosition godoc
func UnmarshalPosition(b []byte) *Position {
	p := new(Position)
	p.bytes = b
	p.level = b[0]
	p.index = big.NewInt(0)
	if len(b) > 1 {
		p.index.SetBytes(b[1:])
	}
	p.setString()
	return p
}

// NewPosition ...
func NewPosition(level uint8, index *big.Int) *Position {
	p := new(Position)
	p.level = level
	p.index = index
	p.setBytes()
	p.setString()
	return p
}

func (p *Position) setBytes() {
	ib := p.index.Bytes()
	p.bytes = make([]byte, 0, 1+len(ib))
	p.bytes = append(p.bytes, p.level)
	p.bytes = append(p.bytes, ib...)
}

func (p *Position) setString() {
	p.str = string(p.bytes)
}

// Level ...
func (p *Position) Level() uint8 {
	return p.level
}

// Index ...
func (p *Position) Index() *big.Int {
	return p.index
}

// Bytes ...
func (p *Position) Bytes() []byte {
	return p.bytes
}

func (p *Position) String() string {
	return p.str
}

// Node type
type Node struct {
	Position *Position
	Data     []byte
}

// UpdateResult type
type UpdateResult struct {
	Nodes     []*Node
	LeafCount *big.Int
	Height    uint8
}
