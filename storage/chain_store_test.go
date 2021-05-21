// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package storage

import (
	"testing"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/stretchr/testify/assert"
)

func TestChainStore(t *testing.T) {
	assert := assert.New(t)
	db := createOnMemoryDB()
	cs := &chainStore{&badgerGetter{db}}

	priv := core.GenerateKey(nil)
	blk := core.NewBlock().SetHeight(10).Sign(priv)

	bcm := core.NewBlockCommit().
		SetHash(blk.Hash()).
		SetMerkleRoot([]byte{1})

	tx := core.NewTransaction().SetNonce(1).Sign(priv)

	txc := core.NewTxCommit().
		SetHash(tx.Hash()).
		SetBlockHash(blk.Hash())

	var err error
	_, err = cs.getBlock(blk.Hash())
	assert.Error(err)
	_, err = cs.getBlockByHeight(blk.Height())
	assert.Error(err)
	_, err = cs.getLastBlock()
	assert.Error(err)
	_, err = cs.getBlockCommit(bcm.Hash())
	assert.Error(err)
	_, err = cs.getTx(tx.Hash())
	assert.Error(err)
	assert.False(cs.hasTx(tx.Hash()))
	_, err = cs.getTxCommit(tx.Hash())
	assert.Error(err)

	updfns := make([]updateFunc, 0)
	updfns = append(updfns, cs.setBlock(blk)...)
	updfns = append(updfns, cs.setBlockHeight(blk.Height()))
	updfns = append(updfns, cs.setBlockCommit(bcm))
	updfns = append(updfns, cs.setTx(tx))
	updfns = append(updfns, cs.setTxCommit(txc))

	updateBadgerDB(db, updfns)

	blk1, err := cs.getBlock(blk.Hash())
	assert.NoError(err)
	assert.Equal(blk.Height(), blk1.Height())

	blk2, err := cs.getBlockByHeight(blk.Height())
	assert.NoError(err)
	assert.Equal(blk.Hash(), blk2.Hash())

	blk3, err := cs.getLastBlock()
	assert.NoError(err)
	assert.Equal(blk.Hash(), blk3.Hash())

	bcm1, err := cs.getBlockCommit(bcm.Hash())
	assert.NoError(err)
	assert.Equal(bcm.MerkleRoot(), bcm1.MerkleRoot())

	tx1, err := cs.getTx(tx.Hash())
	assert.NoError(err)
	assert.Equal(tx.Nonce(), tx1.Nonce())

	assert.True(cs.hasTx(tx.Hash()))

	txc1, err := cs.getTxCommit(tx.Hash())
	assert.NoError(err)
	assert.Equal(txc.BlockHash(), txc1.BlockHash())
}
