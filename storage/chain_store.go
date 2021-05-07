package storage

import (
	"bytes"
	"encoding/binary"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/dgraph-io/badger/v3"
)

type ChainStore struct {
	db *badger.DB
}

func (cs *ChainStore) LoadBlock(hash []byte) (*core.Block, error) {
	b, err := getValue(cs.db, concatBytes([]byte{colBlockByHash}, hash))
	if err != nil {
		return nil, err
	}
	blk := core.NewBlock()
	return blk, blk.Unmarshal(b)
}

func (cs *ChainStore) LoadLastBlock() (*core.Block, error) {
	height, err := cs.LoadBlockHeight()
	if err != nil {
		return nil, err
	}
	return cs.LoadBlockByHeight(height)
}

func (cs *ChainStore) LoadBlockHeight() (uint64, error) {
	b, err := getValue(cs.db, []byte{colBlockHeight})
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(b), nil
}

func (cs *ChainStore) LoadBlockByHeight(height uint64) (*core.Block, error) {
	hash, err := getValue(
		cs.db, concatBytes([]byte{colBlockByHeight}, uint64BEBytes(height)),
	)
	if err != nil {
		return nil, err
	}
	return cs.LoadBlock(hash)
}

func (cs *ChainStore) LoadBlockCommit(hash []byte) (*core.BlockCommit, error) {
	b, err := getValue(cs.db, concatBytes([]byte{colBlockCommitByHash}, hash))
	if err != nil {
		return nil, err
	}
	bcm := core.NewBlockCommit()
	return bcm, bcm.Unmarshal(b)
}

func (cs *ChainStore) LoadTx(hash []byte) (*core.Transaction, error) {
	b, err := getValue(cs.db, concatBytes([]byte{colTxByHash}, hash))
	if err != nil {
		return nil, err
	}
	tx := core.NewTransaction()
	return tx, tx.Unmarshal(b)
}

func (cs *ChainStore) HasTx(hash []byte) bool {
	return hasKey(cs.db, concatBytes([]byte{colTxByHash}, hash))
}

func (cs *ChainStore) LoadTxCommit(hash []byte) (*core.TxCommit, error) {
	val, err := getValue(cs.db, concatBytes([]byte{colTxCommitByHash}, hash))
	if err != nil {
		return nil, err
	}
	txc := core.NewTxCommit()
	return txc, txc.Unmarshal(val)
}

func (cs *ChainStore) storeBlockHeight(height uint64) updateFunc {
	return func(txn *badger.Txn) error {
		return txn.Set([]byte{colBlockHeight}, uint64BEBytes(height))
	}
}

func (cs *ChainStore) storeBlock(block *core.Block) updateFunc {
	return func(txn *badger.Txn) error {
		val, err := block.Marshal()
		if err != nil {
			return err
		}
		err = txn.Set(
			concatBytes([]byte{colBlockByHash}, block.Hash()), val,
		)
		if err != nil {
			return err
		}
		return txn.Set(
			concatBytes([]byte{colBlockByHeight}, uint64BEBytes(block.Height())),
			block.Hash(),
		)
	}
}

func (cs *ChainStore) storeBlockCommit(bcm *core.BlockCommit) updateFunc {
	return func(txn *badger.Txn) error {
		val, err := bcm.Marshal()
		if err != nil {
			return err
		}
		return txn.Set(
			concatBytes([]byte{colBlockCommitByHash}, bcm.Hash()), val,
		)
	}
}

func (cs *ChainStore) storeTx(tx *core.Transaction) updateFunc {
	return func(txn *badger.Txn) error {
		val, err := tx.Marshal()
		if err != nil {
			return err
		}
		return txn.Set(
			concatBytes([]byte{colTxByHash}, tx.Hash()), val,
		)
	}
}

func (cs *ChainStore) storeTxCommit(txc *core.TxCommit) updateFunc {
	return func(txn *badger.Txn) error {
		val, err := txc.Marshal()
		if err != nil {
			return err
		}
		return txn.Set(
			concatBytes([]byte{colTxCommitByHash}, txc.Hash()), val,
		)
	}
}

func uint64BEBytes(val uint64) []byte {
	buf := bytes.NewBuffer(nil)
	binary.Write(buf, binary.BigEndian, val)
	return buf.Bytes()
}
