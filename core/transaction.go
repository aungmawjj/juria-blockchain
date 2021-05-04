// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package core

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/aungmawjj/juria-blockchain/core/core_pb"
	"github.com/aungmawjj/juria-blockchain/util"
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
	data   *core_pb.Transaction
	sender *PublicKey
}

func NewTransaction() *Transaction {
	return &Transaction{
		data: new(core_pb.Transaction),
	}
}

// Sum returns sha3 sum of transaction
func (tx *Transaction) Sum() []byte {
	h := sha3.New256()
	binary.Write(h, util.ByteOrder, tx.data.Nonce)
	h.Write(tx.data.Sender)
	h.Write(tx.data.CodeAddr)
	h.Write(tx.data.Input)
	return h.Sum(nil)
}

// Validate transaction
func (tx *Transaction) Validate() error {
	if tx.data == nil {
		return ErrNilTx
	}
	if !bytes.Equal(tx.Sum(), tx.Hash()) {
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

func (tx *Transaction) setData(data *core_pb.Transaction) *Transaction {
	tx.data = data
	tx.sender, _ = NewPublicKey(tx.data.Sender)
	return tx
}

func (tx *Transaction) SetNonce(val uint64) *Transaction {
	tx.data.Nonce = val
	return tx
}

func (tx *Transaction) SetCodeAddr(val []byte) *Transaction {
	tx.data.CodeAddr = val
	return tx
}

func (tx *Transaction) SetInput(val []byte) *Transaction {
	tx.data.Input = val
	return tx
}

func (tx *Transaction) Sign(priv *PrivateKey) *Transaction {
	tx.sender = priv.PublicKey()
	tx.data.Sender = priv.PublicKey().key
	tx.data.Hash = tx.Sum()
	tx.data.Signature = priv.Sign(tx.data.Hash).data.Value
	return tx
}

func (tx *Transaction) Hash() []byte       { return tx.data.Hash }
func (tx *Transaction) Nonce() uint64      { return tx.data.Nonce }
func (tx *Transaction) Sender() *PublicKey { return tx.sender }
func (tx *Transaction) CodeAddr() []byte   { return tx.data.CodeAddr }
func (tx *Transaction) Input() []byte      { return tx.data.Input }

// Marshal encodes transaction as bytes
func (tx *Transaction) Marshal() ([]byte, error) {
	return proto.Marshal(tx.data)
}

// UnmarshalTransaction decodes transaction from bytes
func (tx *Transaction) Unmarshal(b []byte) error {
	data := new(core_pb.Transaction)
	if err := proto.Unmarshal(b, data); err != nil {
		return err
	}
	tx.setData(data)
	return nil
}

type TxList []*Transaction

func NewTxList() *TxList {
	return new(TxList)
}

// UnmarshalTxList decodes tx list from bytes
func (txs *TxList) Unmarshal(b []byte) error {
	data := new(core_pb.TxList)
	if err := proto.Unmarshal(b, data); err != nil {
		return err
	}
	*txs = make([]*Transaction, len(data.List))
	for i, txData := range data.List {
		(*txs)[i] = NewTransaction().setData(txData)
	}
	return nil
}

// Marshal encodes tx list as bytes
func (txs *TxList) Marshal() ([]byte, error) {
	data := new(core_pb.TxList)
	data.List = make([]*core_pb.Transaction, len(*txs))
	for i, tx := range *txs {
		data.List[i] = tx.data
	}
	return proto.Marshal(data)
}

type HashList [][]byte

func NewHashList() *HashList {
	return new(HashList)
}

func (hl *HashList) Unmarshal(b []byte) error {
	data := new(core_pb.HashList)
	if err := proto.Unmarshal(b, data); err != nil {
		return err
	}
	*hl = data.List
	return nil
}

func (hl *HashList) Marshal() ([]byte, error) {
	data := new(core_pb.HashList)
	data.List = *hl
	return proto.Marshal(data)
}
