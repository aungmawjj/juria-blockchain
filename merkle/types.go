// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package merkle

import "math/big"

// Position of a node in the tree
type Position struct {
	Level uint8
	Index *big.Int
}

// Bytes binary value of a position
func (p *Position) Bytes() []byte {
	ib := p.Index.Bytes()
	b := make([]byte, 0, 1+len(ib))
	b = append(b, p.Level)
	b = append(b, ib...)
	return b
}
