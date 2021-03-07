package merkle

import "math/big"

// Position of a node in the tree
type Position struct {
	Level uint8
	Index *big.Int
}

// Bytes binary value of a position
func (p *Position) Bytes() []byte {
	return append([]byte{p.Level}, p.Index.Bytes()...)
}
