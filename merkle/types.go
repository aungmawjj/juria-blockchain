// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package merkle

import (
	"crypto"
	"math/big"
)

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

// Block type
type Block struct {
	hashFunc       crypto.Hash
	tc             *TreeCalc
	store          Store
	parentPosition *Position
	nodes          []*Node
}

// NewBlock ...
func NewBlock(h crypto.Hash, tc *TreeCalc, store Store, parentPosition *Position) *Block {
	if parentPosition.Level() < 1 {
		panic("parent level should be at least 1")
	}
	return &Block{
		hashFunc:       h,
		tc:             tc,
		store:          store,
		parentPosition: parentPosition,
		nodes:          make([]*Node, int(tc.BranchFactor())),
	}
}

// ParentPosition ...
func (b *Block) ParentPosition() *Position {
	return b.parentPosition
}

// Load ...
func (b *Block) Load() *Block {
	offset := b.tc.FirstNodeOfBlock(b.parentPosition.Index())
	for i := range b.nodes {
		index := big.NewInt(0).Add(offset, big.NewInt(int64(i)))
		p := NewPosition(b.parentPosition.level-1, index)

		if data := b.store.GetNode(p); data != nil {
			b.nodes[i] = &Node{
				Position: p,
				Data:     b.store.GetNode(p),
			}
		}
	}
	return b
}

// SetNode ...
func (b *Block) SetNode(n *Node) *Block {
	b.nodes[b.tc.NodeIndexInBlock(n.Position.Index())] = n
	return b
}

// MakeParent ...
func (b *Block) MakeParent() *Node {
	return &Node{
		Position: b.parentPosition,
		Data:     b.Sum(),
	}
}

// Sum ...
func (b *Block) Sum() []byte {
	if b.IsEmpty() {
		return nil
	}
	h := b.hashFunc.New()
	for _, n := range b.nodes {
		if n != nil {
			h.Write(n.Data)
		}
	}
	return h.Sum(nil)
}

// IsEmpty ...
func (b *Block) IsEmpty() bool {
	for _, n := range b.nodes {
		if n != nil {
			return false
		}
	}
	return true
}

// UpdateResult type
type UpdateResult struct {
	LeafCount *big.Int
	Height    uint8
	Nodes     []*Node
}
