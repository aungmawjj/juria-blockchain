package storage

import (
	"math/big"
	"testing"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/merkle"
	"github.com/stretchr/testify/assert"
)

func newTestStorage() *Storage {
	return NewStorage(createOnMemoryDB(), merkle.TreeOptions{
		BranchFactor: 8,
		HashFunc:     hashFunc,
	})
}

func TestStorage_StateZero(t *testing.T) {
	assert := assert.New(t)

	strg := newTestStorage()
	assert.Nil(strg.GetMerkleRoot())
	_, err := strg.GetBlockHeight()
	assert.Error(err)
	assert.Nil(strg.GetState([]byte("some key")))
}

func TestStorage_Commit(t *testing.T) {
	assert := assert.New(t)

	strg := newTestStorage()
	b0 := core.NewBlock().SetHeight(0)
	b0.SetHash(b0.Sum())
	data := &CommitData{
		Block: b0,
		StateChanges: []*core.StateChange{
			core.NewStateChange().SetKey([]byte{1}).SetValue([]byte{10}),
			core.NewStateChange().SetKey([]byte{2}).SetValue([]byte{20}),
		},
	}
	err := strg.Commit(data)
	assert.NoError(err)

	blkHeight, err := strg.GetBlockHeight()
	assert.NoError(err)
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

	b1 := core.NewBlock().
		SetHeight(1).
		SetParentHash(b0.Hash()).
		SetMerkleRoot(strg.GetMerkleRoot())
	b1.SetHash(b1.Sum())

	tx1 := core.NewTransaction().SetNonce(1)
	tx1.SetHash(tx1.Sum())
	tx2 := core.NewTransaction().SetNonce(2)
	tx2.SetHash(tx2.Sum())

	txc1 := core.NewTxCommit().SetHash(tx1.Hash())
	txc2 := core.NewTxCommit().SetHash(tx2.Hash())

	data = &CommitData{
		Block:        b1,
		Transactions: []*core.Transaction{tx1, tx2},
		TxCommits:    []*core.TxCommit{txc1, txc2},
		StateChanges: []*core.StateChange{
			core.NewStateChange().SetKey([]byte{1}).SetValue([]byte{20}),

			// new key, leaf index -> 2, increase leaf count -> 3
			core.NewStateChange().SetKey([]byte{5}).SetValue([]byte{30}),

			// new key, leaf index -> 3, increase leaf count -> 4
			core.NewStateChange().SetKey([]byte{3}).SetValue([]byte{50}),
		},
	}
	err = strg.Commit(data)
	assert.NoError(err)

	blkHeight, err = strg.GetBlockHeight()
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
	assert.Equal(big.NewInt(2).Bytes(), bcm.StateChanges()[1].TreeIndex())
	assert.Equal(big.NewInt(3).Bytes(), bcm.StateChanges()[2].TreeIndex())
	assert.Equal(big.NewInt(4).Bytes(), bcm.LeafCount())

	h.Write(strg.stateStore.sumStateValue([]byte{20}))
	h.Write(strg.stateStore.sumStateValue([]byte{20}))
	h.Write(strg.stateStore.sumStateValue([]byte{30}))
	h.Write(strg.stateStore.sumStateValue([]byte{50}))
	mroot = h.Sum(nil)
	h.Reset()
	assert.Equal(mroot, bcm.MerkleRoot())
	assert.Equal(bcm.MerkleRoot(), strg.GetMerkleRoot())

	assert.Equal([]byte{30}, strg.GetState([]byte{5}))
}
