// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package chaincode

type ReadContext interface {
	Input() []byte
	// GetState returns state value verified with merkle root
	GetState(key []byte) ([]byte, error)
}

type WriteContext interface {
	Sender() []byte
	BlockHash() []byte
	BlockHeight() uint64
	Input() []byte

	GetState(key []byte) []byte
	SetState(key, value []byte)
}

// all chaincodes implements ChainCode interface
type ChainCode interface {
	// called when chaincode is deployed
	Init(wc WriteContext) error

	Invoke(wc WriteContext) error

	Query(rc ReadContext) ([]byte, error)
}
