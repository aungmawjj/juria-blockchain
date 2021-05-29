// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package core

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTransaction(t *testing.T) {
	privKey := GenerateKey(nil)

	nonce := time.Now().UnixNano()

	tx := NewTransaction().
		SetNonce(nonce).
		SetCodeAddr([]byte{1}).
		SetInput([]byte{2}).
		Sign(privKey)

	assert := assert.New(t)

	assert.Equal(nonce, tx.Nonce())
	assert.Equal([]byte{1}, tx.CodeAddr())
	assert.Equal([]byte{2}, tx.Input())
	assert.Equal(privKey.PublicKey(), tx.Sender())
	assert.Equal(privKey.PublicKey().Bytes(), tx.data.Sender)

	assert.NoError(tx.Validate())

	b, err := tx.Marshal()
	assert.NoError(err)

	tx = NewTransaction()
	err = tx.Unmarshal(b)
	assert.NoError(err)

	assert.NoError(tx.Validate())

	b, err = json.Marshal(tx)
	assert.NoError(err)

	tx = NewTransaction()
	err = json.Unmarshal(b, tx)
	assert.NoError(err)

	assert.NoError(tx.Validate())
}

func TestTxList(t *testing.T) {
	privKey := GenerateKey(nil)

	tx1 := NewTransaction().
		SetNonce(time.Now().UnixNano()).
		SetCodeAddr([]byte{1}).
		SetInput([]byte{2}).
		Sign(privKey)

	tx2 := NewTransaction().
		SetNonce(time.Now().UnixNano()).
		SetCodeAddr([]byte{2}).
		SetInput([]byte{2}).
		Sign(privKey)

	assert := assert.New(t)

	txs := &TxList{tx1, tx2}
	b, err := txs.Marshal()
	assert.NoError(err)

	txs = NewTxList()
	err = txs.Unmarshal(b)
	assert.NoError(err)

	assert.Equal(2, len(*txs))
	assert.Equal(tx1.Sum(), (*txs)[0].Sum())
	assert.Equal(tx2.Sum(), (*txs)[1].Sum())
}
