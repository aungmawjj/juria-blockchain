package storage

import (
	"testing"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/stretchr/testify/assert"
)

func TestChainStore(t *testing.T) {
	assert := assert.New(t)
	db := createOnMemoryDB()
	cs := &ChainStore{db}

	blk := core.NewBlock().
		SetHeight(10)
	blk.SetHash(blk.Sum())

	bcm := core.NewBlockCommit().
		SetHash(blk.Hash()).
		SetStateRoot([]byte{1})

	tx := core.NewTransaction().
		SetNonce(1)
	tx.SetHash(tx.Sum())

	txc := core.NewTxCommit().
		SetHash(tx.Hash()).
		SetBlockHash(blk.Hash())

	var err error
	_, err = cs.LoadBlock(blk.Hash())
	assert.Error(err)
	_, err = cs.LoadBlockByHeight(blk.Height())
	assert.Error(err)
	_, err = cs.LoadLastBlock()
	assert.Error(err)
	_, err = cs.LoadBlockCommit(bcm.Hash())
	assert.Error(err)
	_, err = cs.LoadTx(tx.Hash())
	assert.Error(err)
	assert.False(cs.HasTx(tx.Hash()))
	_, err = cs.LoadTxCommit(tx.Hash())
	assert.Error(err)

	updfns := make([]updateFunc, 0)
	updfns = append(updfns, cs.storeBlock(blk))
	updfns = append(updfns, cs.storeBlockHeight(blk.Height()))
	updfns = append(updfns, cs.storeBlockCommit(bcm))
	updfns = append(updfns, cs.storeTx(tx))
	updfns = append(updfns, cs.storeTxCommit(txc))

	updateDB(db, updfns)

	blk1, err := cs.LoadBlock(blk.Hash())
	assert.NoError(err)
	assert.Equal(blk.Height(), blk1.Height())

	blk2, err := cs.LoadBlockByHeight(blk.Height())
	assert.NoError(err)
	assert.Equal(blk.Hash(), blk2.Hash())

	blk3, err := cs.LoadLastBlock()
	assert.NoError(err)
	assert.Equal(blk.Hash(), blk3.Hash())

	bcm1, err := cs.LoadBlockCommit(bcm.Hash())
	assert.NoError(err)
	assert.Equal(bcm.StateRoot(), bcm1.StateRoot())

	tx1, err := cs.LoadTx(tx.Hash())
	assert.NoError(err)
	assert.Equal(tx.Nonce(), tx1.Nonce())

	assert.True(cs.HasTx(tx.Hash()))

	txc1, err := cs.LoadTxCommit(tx.Hash())
	assert.NoError(err)
	assert.Equal(txc.BlockHash(), txc1.BlockHash())
}
