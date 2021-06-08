// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package core

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"

	"github.com/aungmawjj/juria-blockchain/core/core_pb"
	"golang.org/x/crypto/sha3"
	"google.golang.org/protobuf/encoding/protojson"
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

var _ json.Unmarshaler = (*Transaction)(nil)
var _ json.Marshaler = (*Transaction)(nil)

func NewTransaction() *Transaction {
	return &Transaction{
		data: new(core_pb.Transaction),
	}
}

// Sum returns sha3 sum of transaction
func (tx *Transaction) Sum() []byte {
	h := sha3.New256()
	binary.Write(h, binary.BigEndian, tx.data.Nonce)
	h.Write(tx.data.Sender)
	h.Write(tx.data.CodeAddr)
	h.Write(tx.data.Input)
	binary.Write(h, binary.BigEndian, tx.data.Expiry)
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

func (tx *Transaction) setData(data *core_pb.Transaction) error {
	tx.data = data
	var err error
	tx.sender, err = NewPublicKey(tx.data.Sender)
	return err
}

func (tx *Transaction) SetNonce(val int64) *Transaction {
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

func (tx *Transaction) SetExpiry(val uint64) *Transaction {
	tx.data.Expiry = val
	return tx
}

func (tx *Transaction) Sign(signer Signer) *Transaction {
	tx.sender = signer.PublicKey()
	tx.data.Sender = signer.PublicKey().key
	tx.data.Hash = tx.Sum()
	tx.data.Signature = signer.Sign(tx.data.Hash).data.Value
	return tx
}

func (tx *Transaction) Hash() []byte       { return tx.data.Hash }
func (tx *Transaction) Nonce() int64       { return tx.data.Nonce }
func (tx *Transaction) Sender() *PublicKey { return tx.sender }
func (tx *Transaction) CodeAddr() []byte   { return tx.data.CodeAddr }
func (tx *Transaction) Input() []byte      { return tx.data.Input }
func (tx *Transaction) Expiry() uint64     { return tx.data.Expiry }

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
	return tx.setData(data)
}

func (tx *Transaction) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(tx.data)
}

func (tx *Transaction) UnmarshalJSON(b []byte) error {
	data := new(core_pb.Transaction)
	if err := protojson.Unmarshal(b, data); err != nil {
		return err
	}
	return tx.setData(data)
}

type TxCommit struct {
	data *core_pb.TxCommit
}

var _ json.Marshaler = (*TxCommit)(nil)
var _ json.Unmarshaler = (*TxCommit)(nil)

func NewTxCommit() *TxCommit {
	return &TxCommit{
		data: new(core_pb.TxCommit),
	}
}

func (txc *TxCommit) Hash() []byte        { return txc.data.Hash }
func (txc *TxCommit) BlockHash() []byte   { return txc.data.BlockHash }
func (txc *TxCommit) BlockHeight() uint64 { return txc.data.BlockHeight }
func (txc *TxCommit) Elapsed() float64    { return txc.data.Elapsed }
func (txc *TxCommit) Error() string       { return txc.data.Error }

func (txc *TxCommit) SetHash(val []byte) *TxCommit {
	txc.data.Hash = val
	return txc
}

func (txc *TxCommit) SetBlockHash(val []byte) *TxCommit {
	txc.data.BlockHash = val
	return txc
}

func (txc *TxCommit) SetBlockHeight(val uint64) *TxCommit {
	txc.data.BlockHeight = val
	return txc
}

func (txc *TxCommit) SetElapsed(val float64) *TxCommit {
	txc.data.Elapsed = val
	return txc
}

func (txc *TxCommit) SetError(val string) *TxCommit {
	txc.data.Error = val
	return txc
}

func (txc *TxCommit) setData(data *core_pb.TxCommit) error {
	txc.data = data
	return nil
}

func (txc *TxCommit) Marshal() ([]byte, error) {
	return proto.Marshal(txc.data)
}

func (txc *TxCommit) Unmarshal(b []byte) error {
	data := new(core_pb.TxCommit)
	if err := proto.Unmarshal(b, data); err != nil {
		return err
	}
	return txc.setData(data)
}

func (txc *TxCommit) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(txc.data)
}

func (txc *TxCommit) UnmarshalJSON(b []byte) error {
	data := new(core_pb.TxCommit)
	if err := protojson.Unmarshal(b, data); err != nil {
		return err
	}
	return txc.setData(data)
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
		tx := NewTransaction()
		if err := tx.setData(txData); err != nil {
			return err
		}
		(*txs)[i] = tx
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
