// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package core

import (
	"bytes"
	"errors"

	core_pb "github.com/aungmawjj/juria-blockchain/core/pb"
	"golang.org/x/crypto/sha3"
	"google.golang.org/protobuf/proto"
)

// errors
var (
	ErrInvalidTxHash = errors.New("invalid tx hash")
	ErrNilTx         = errors.New("nil tx")
)

// Transaction type
type Transaction struct {
	data *core_pb.Transaction
}

// newTransaction creates tx from pb data
func newTransaction(data *core_pb.Transaction) (*Transaction, error) {
	if data == nil {
		return nil, ErrNilTx
	}
	return &Transaction{data}, nil
}

// UnmarshalTransaction decodes transaction from bytes
func UnmarshalTransaction(b []byte) (*Transaction, error) {
	data := new(core_pb.Transaction)
	if err := proto.Unmarshal(b, data); err != nil {
		return nil, err
	}
	return newTransaction(data)
}

// Marshal encodes transaction as bytes
func (tx *Transaction) Marshal() ([]byte, error) {
	return proto.Marshal(tx.data)
}

// Sum returns sha3 sum of transaction
func (tx *Transaction) Sum() []byte {
	h := sha3.New256()
	h.Write(uint64ToBytes(tx.data.Nonce))
	h.Write(tx.data.Sender)
	h.Write(tx.data.CodeAddr)
	h.Write(tx.data.Data)
	return h.Sum(nil)
}

// Validate transaction
func (tx *Transaction) Validate() error {
	if !bytes.Equal(tx.Sum(), tx.data.Hash) {
		return ErrInvalidTxHash
	}
	sig, err := newSignature(&core_pb.Signature{
		PubKey: tx.data.Sender,
		Value:  tx.data.Signature,
	})
	if err != nil {
		return err
	}
	if !sig.Verify(tx.data.Hash) {
		return ErrInvalidSig
	}
	return nil
}

// UnmarshalTxList decodes tx list from bytes
func UnmarshalTxList(b []byte) ([]*Transaction, error) {
	data := new(core_pb.TxList)
	if err := proto.Unmarshal(b, data); err != nil {
		return nil, err
	}
	txs := make([]*Transaction, len(data.List))
	for i, txData := range data.List {
		tx, err := newTransaction(txData)
		if err != nil {
			return nil, err
		}
		txs[i] = tx
	}
	return txs, nil
}

// MarshalTxList encodes tx list as bytes
func MarshalTxList(txs []*Transaction) ([]byte, error) {
	data := new(core_pb.TxList)
	data.List = make([]*core_pb.Transaction, len(txs))
	for i, tx := range txs {
		data.List[i] = tx.data
	}
	return proto.Marshal(data)
}
