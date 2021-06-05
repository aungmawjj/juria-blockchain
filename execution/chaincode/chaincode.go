// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package chaincode

type CallContext interface {
	Sender() []byte
	BlockHash() []byte
	BlockHeight() uint64
	Input() []byte

	GetState(key []byte) []byte
	SetState(key, value []byte)
}

// all chaincodes implements Chaincode interface
type Chaincode interface {
	// called when chaincode is deployed
	Init(ctx CallContext) error

	Invoke(ctx CallContext) error

	Query(ctx CallContext) ([]byte, error)
}
