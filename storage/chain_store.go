package storage

import (
	"bytes"
	"encoding/binary"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/util"
	"github.com/dgraph-io/badger"
)

type ChainStore struct {
	db *badger.DB
}

func (cs *ChainStore) LoadBlock(hash []byte) (*core.Block, error) {
	b, err := getValue(cs.db, util.ConcatBytes([]byte{colBlockByHash}, hash))
	if err != nil {
		return nil, err
	}
	blk := core.NewBlock()
	return blk, blk.Unmarshal(b)
}

func (cs *ChainStore) LoadBlockHeight() (uint64, error) {
	b, err := getValue(cs.db, []byte{colBlockHeight})
	if err != nil {
		return 0, err
	}
	return util.ByteOrder.Uint64(b), nil
}

func (cs *ChainStore) LoadBlockHashByHeight(height uint64) ([]byte, error) {
	key := bytes.NewBuffer(nil)
	key.WriteByte(colBlockByHeight)
	binary.Write(key, util.ByteOrder, height)
	return getValue(cs.db, key.Bytes())
}

func (cs *ChainStore) LoadLastBlock() (*core.Block, error) {
	height, err := cs.LoadBlockHeight()
	if err != nil {
		return nil, err
	}
	hash, err := cs.LoadBlockHashByHeight(height)
	if err != nil {
		return nil, err
	}
	return cs.LoadBlock(hash)
}

func (cs *ChainStore) LoadBlockCommit(hash []byte) (*core.BlockCommit, error) {
	b, err := getValue(cs.db, util.ConcatBytes([]byte{colBlockCommitByHash}, hash))
	if err != nil {
		return nil, err
	}
	bcm := core.NewBlockCommit()
	return bcm, bcm.Unmarshal(b)
}

func (cs *ChainStore) LoadTx(hash []byte) (*core.Transaction, error) {
	b, err := getValue(cs.db, util.ConcatBytes([]byte{colTxByHash}, hash))
	if err != nil {
		return nil, err
	}
	tx := core.NewTransaction()
	return tx, tx.Unmarshal(b)
}

func (cs *ChainStore) HasTx(hash []byte) bool {
	return hasKey(cs.db, util.ConcatBytes([]byte{colTxByHash}, hash))
}
