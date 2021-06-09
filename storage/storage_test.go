// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package storage

import (
	"math/big"
	"testing"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/stretchr/testify/assert"
)

func newTestStorage() *Storage {
	return New(createOnMemoryDB(), DefaultConfig)
}

func TestStorage_StateZero(t *testing.T) {
	assert := assert.New(t)

	strg := newTestStorage()
	assert.Nil(strg.GetMerkleRoot())
	_, err := strg.GetLastBlock()
	assert.Error(err)
	assert.Nil(strg.GetState([]byte("some key")))
}

func TestStorage_Commit(t *testing.T) {
	assert := assert.New(t)

	strg := newTestStorage()
	priv := core.GenerateKey(nil)
	b0 := core.NewBlock().SetHeight(0).Sign(priv)
	bcmInput := core.NewBlockCommit().
		SetHash(b0.Hash()).
		SetStateChanges([]*core.StateChange{
			core.NewStateChange().SetKey([]byte{1}).SetValue([]byte{10}),
			core.NewStateChange().SetKey([]byte{2}).SetValue([]byte{20}),
		})
	data := &CommitData{
		Block:       b0,
		QC:          core.NewQuorumCert(),
		BlockCommit: bcmInput,
	}
	err := strg.Commit(data)
	assert.NoError(err)

	blkHeight := strg.GetBlockHeight()
	assert.EqualValues(0, blkHeight)

	blk, err := strg.GetLastBlock()
	assert.NoError(err)
	assert.Equal(b0.Hash(), blk.Hash())

	blk, err = strg.GetBlock(b0.Hash())
	assert.NoError(err)
	assert.Equal(b0.Hash(), blk.Hash())

	blk, err = strg.GetBlockByHeight(0)
	assert.NoError(err)
	assert.Equal(b0.Hash(), blk.Hash())

	bcm, err := strg.GetBlockCommit(b0.Hash())
	assert.NoError(err)
	assert.Equal([]byte{0}, bcm.StateChanges()[0].TreeIndex())
	assert.Equal(big.NewInt(1).Bytes(), bcm.StateChanges()[1].TreeIndex())
	assert.Equal(big.NewInt(2).Bytes(), bcm.LeafCount())

	h := hashFunc.New()
	h.Write(strg.stateStore.sumStateValue([]byte{10}))
	h.Write(strg.stateStore.sumStateValue([]byte{20}))
	mroot := h.Sum(nil)
	h.Reset()
	assert.Equal(mroot, bcm.MerkleRoot())
	assert.Equal(bcm.MerkleRoot(), strg.GetMerkleRoot())

	assert.Equal([]byte{10}, strg.GetState([]byte{1}))
	assert.Equal([]byte{20}, strg.GetState([]byte{2}))

	qc := core.NewQuorumCert().Build([]*core.Vote{b0.ProposerVote()})
	b1 := core.NewBlock().
		SetHeight(1).
		SetQuorumCert(qc).
		SetParentHash(b0.Hash()).
		SetMerkleRoot(strg.GetMerkleRoot()).
		Sign(priv)

	tx1 := core.NewTransaction().SetNonce(1).Sign(priv)
	tx2 := core.NewTransaction().SetNonce(2).Sign(priv)

	txc1 := core.NewTxCommit().SetHash(tx1.Hash())
	txc2 := core.NewTxCommit().SetHash(tx2.Hash())

	bcmInput = core.NewBlockCommit().
		SetHash(b1.Hash()).
		SetStateChanges([]*core.StateChange{
			core.NewStateChange().SetKey([]byte{1}).SetValue([]byte{20}),

			// new key, leaf index -> 3, increase leaf count -> 4
			core.NewStateChange().SetKey([]byte{5}).SetValue([]byte{50}),

			// new key, leaf index -> 2, increase leaf count -> 3
			core.NewStateChange().SetKey([]byte{3}).SetValue([]byte{30}),
		})
	data = &CommitData{
		Block:        b1,
		QC:           core.NewQuorumCert(),
		Transactions: []*core.Transaction{tx1, tx2},
		TxCommits:    []*core.TxCommit{txc1, txc2},
		BlockCommit:  bcmInput,
	}
	err = strg.Commit(data)
	assert.NoError(err)

	blkHeight = strg.GetBlockHeight()
	assert.NoError(err)
	assert.EqualValues(1, blkHeight)

	blk, err = strg.GetLastBlock()
	assert.NoError(err)
	assert.Equal(b1.Hash(), blk.Hash())

	tx, err := strg.GetTx(tx1.Hash())
	assert.NoError(err)
	assert.Equal(tx1.Nonce(), tx.Nonce())

	assert.True(strg.HasTx(tx2.Hash()))

	txc, err := strg.GetTxCommit(tx2.Hash())
	assert.NoError(err)
	assert.Equal(txc2.Hash(), txc.Hash())

	bcm, err = strg.GetBlockCommit(b1.Hash())
	assert.NoError(err)
	assert.Equal([]byte{0}, bcm.StateChanges()[0].PrevTreeIndex())
	assert.Equal(big.NewInt(3).Bytes(), bcm.StateChanges()[1].TreeIndex())
	assert.Equal(big.NewInt(2).Bytes(), bcm.StateChanges()[2].TreeIndex())
	assert.Equal(big.NewInt(4).Bytes(), bcm.LeafCount())

	h.Write(strg.stateStore.sumStateValue([]byte{20}))
	h.Write(strg.stateStore.sumStateValue([]byte{20}))
	h.Write(strg.stateStore.sumStateValue([]byte{30}))
	h.Write(strg.stateStore.sumStateValue([]byte{50}))
	mroot = h.Sum(nil)
	h.Reset()
	assert.Equal(mroot, bcm.MerkleRoot())
	assert.Equal(bcm.MerkleRoot(), strg.GetMerkleRoot())

	assert.Equal([]byte{50}, strg.GetState([]byte{5}))
	var value []byte
	assert.NotPanics(func() {
		value = strg.VerifyState([]byte{5})
	})
	assert.Equal([]byte{50}, value)

	assert.NotPanics(func() {
		// non existing state value
		value = strg.VerifyState([]byte{10})
	})
	assert.Nil(value)

	// tampering state value
	updFn := strg.stateStore.setState([]byte{5}, []byte{100})
	updateBadgerDB(strg.db, []updateFunc{updFn})

	// should panic
	assert.Panics(func() {
		value = strg.VerifyState([]byte{5})
	})
	assert.Nil(value)
}
