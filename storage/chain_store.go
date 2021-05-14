package storage

import (
	"bytes"
	"encoding/binary"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/dgraph-io/badger/v3"
)

type chainStore struct {
	db *badger.DB
}

func (cs *chainStore) getLastBlock() (*core.Block, error) {
	height, err := cs.getBlockHeight()
	if err != nil {
		return nil, err
	}
	return cs.getBlockByHeight(height)
}

func (cs *chainStore) getBlockHeight() (uint64, error) {
	b, err := getValue(cs.db, []byte{colBlockHeight})
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(b), nil
}

func (cs *chainStore) getBlockByHeight(height uint64) (*core.Block, error) {
	hash, err := cs.getBlockHashByHeight(height)
	if err != nil {
		return nil, err
	}
	return cs.getBlock(hash)
}

func (cs *chainStore) getBlockHashByHeight(height uint64) ([]byte, error) {
	return getValue(
		cs.db, concatBytes([]byte{colBlockHashByHeight}, uint64BEBytes(height)),
	)
}

func (cs *chainStore) getBlock(hash []byte) (*core.Block, error) {
	b, err := getValue(cs.db, concatBytes([]byte{colBlockByHash}, hash))
	if err != nil {
		return nil, err
	}
	blk := core.NewBlock()
	return blk, blk.Unmarshal(b)
}

func (cs *chainStore) getBlockCommit(hash []byte) (*core.BlockCommit, error) {
	b, err := getValue(cs.db, concatBytes([]byte{colBlockCommitByHash}, hash))
	if err != nil {
		return nil, err
	}
	bcm := core.NewBlockCommit()
	return bcm, bcm.Unmarshal(b)
}

func (cs *chainStore) getTx(hash []byte) (*core.Transaction, error) {
	b, err := getValue(cs.db, concatBytes([]byte{colTxByHash}, hash))
	if err != nil {
		return nil, err
	}
	tx := core.NewTransaction()
	return tx, tx.Unmarshal(b)
}

func (cs *chainStore) hasTx(hash []byte) bool {
	return hasKey(cs.db, concatBytes([]byte{colTxByHash}, hash))
}

func (cs *chainStore) getTxCommit(hash []byte) (*core.TxCommit, error) {
	val, err := getValue(cs.db, concatBytes([]byte{colTxCommitByHash}, hash))
	if err != nil {
		return nil, err
	}
	txc := core.NewTxCommit()
	return txc, txc.Unmarshal(val)
}

func (cs *chainStore) setBlockHeight(height uint64) updateFunc {
	return func(setter setter) error {
		return setter.Set([]byte{colBlockHeight}, uint64BEBytes(height))
	}
}

func (cs *chainStore) setBlock(blk *core.Block) []updateFunc {
	ret := make([]updateFunc, 0)
	ret = append(ret, cs.setBlockByHash(blk))
	ret = append(ret, cs.setBlockHashByHeight(blk))
	return ret
}

func (cs *chainStore) setBlockByHash(blk *core.Block) updateFunc {
	return func(setter setter) error {
		val, err := blk.Marshal()
		if err != nil {
			return err
		}
		return setter.Set(
			concatBytes([]byte{colBlockByHash}, blk.Hash()), val,
		)
	}
}

func (cs *chainStore) setBlockHashByHeight(blk *core.Block) updateFunc {
	return func(setter setter) error {
		return setter.Set(
			concatBytes([]byte{colBlockHashByHeight}, uint64BEBytes(blk.Height())),
			blk.Hash(),
		)
	}
}

func (cs *chainStore) setBlockCommit(bcm *core.BlockCommit) updateFunc {
	return func(setter setter) error {
		val, err := bcm.Marshal()
		if err != nil {
			return err
		}
		return setter.Set(
			concatBytes([]byte{colBlockCommitByHash}, bcm.Hash()), val,
		)
	}
}

func (cs *chainStore) setTx(tx *core.Transaction) updateFunc {
	return func(setter setter) error {
		val, err := tx.Marshal()
		if err != nil {
			return err
		}
		return setter.Set(
			concatBytes([]byte{colTxByHash}, tx.Hash()), val,
		)
	}
}

func (cs *chainStore) setTxCommit(txc *core.TxCommit) updateFunc {
	return func(setter setter) error {
		val, err := txc.Marshal()
		if err != nil {
			return err
		}
		return setter.Set(
			concatBytes([]byte{colTxCommitByHash}, txc.Hash()), val,
		)
	}
}

func uint64BEBytes(val uint64) []byte {
	buf := bytes.NewBuffer(nil)
	binary.Write(buf, binary.BigEndian, val)
	return buf.Bytes()
}
