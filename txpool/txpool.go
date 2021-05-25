// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package txpool

import (
	"bytes"
	"errors"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/emitter"
	"github.com/aungmawjj/juria-blockchain/logger"
)

type Status struct {
	Total   int `json:"total"`
	Pending int `json:"pending"`
	Queue   int `json:"queue"`
}

type Storage interface {
	HasTx(hash []byte) bool
}

type MsgService interface {
	SubscribeTxList(buffer int) *emitter.Subscription
	BroadcastTxList(txList *core.TxList) error
	RequestTxList(pubKey *core.PublicKey, hashes [][]byte) (*core.TxList, error)
}

var (
	ErrOldTx = errors.New("transaction already executed")
)

type TxPool struct {
	storage Storage
	msgSvc  MsgService

	store       *txStore
	broadcaster *broadcaster
}

func New(storage Storage, msgSvc MsgService) *TxPool {
	pool := &TxPool{
		storage:     storage,
		msgSvc:      msgSvc,
		store:       newTxStore(),
		broadcaster: newBroadcaster(msgSvc),
	}
	go pool.subscribeTxs()
	return pool
}

func (pool *TxPool) SubmitTx(tx *core.Transaction) error {
	return pool.submitTx(tx)
}

func (pool *TxPool) SyncTxs(peer *core.PublicKey, hashes [][]byte) error {
	return pool.syncTxs(peer, hashes)
}

func (pool *TxPool) VerifyProposalTxs(hashes [][]byte) error {
	return pool.verifyProposalTxs(hashes)
}

func (pool *TxPool) PopTxsFromQueue(max int) [][]byte {
	return pool.store.popTxsFromQueue(max)
}

func (pool *TxPool) PutTxsToQueue(hashes [][]byte) {
	pool.store.putTxsToQueue(hashes)
}

func (pool *TxPool) SetTxsPending(hashes [][]byte) {
	pool.store.setTxsPending(hashes)
}

func (pool *TxPool) GetTxsToExecute(hashes [][]byte) ([]*core.Transaction, [][]byte) {
	return pool.getTxsToExecute(hashes)
}

func (pool *TxPool) RemoveTxs(hashes [][]byte) {
	pool.store.removeTxs(hashes)
}

func (pool *TxPool) GetTx(hash []byte) *core.Transaction {
	return pool.store.getTx(hash)
}

func (pool *TxPool) GetStatus() Status {
	return pool.store.getStatus()
}

func (pool *TxPool) subscribeTxs() {
	sub := pool.msgSvc.SubscribeTxList(100)
	for e := range sub.Events() {
		txList := e.(*core.TxList)
		if err := pool.addTxList(txList); err != nil {
			logger.I().Warnw("add tx list failed", "error", err)
		}
	}
}

func (pool *TxPool) submitTx(tx *core.Transaction) error {
	if err := tx.Validate(); err != nil {
		return err
	}
	if pool.storage.HasTx(tx.Hash()) {
		return ErrOldTx
	}
	pool.store.addNewTx(tx)
	pool.broadcaster.queue <- tx
	return nil
}

func (pool *TxPool) addTxList(txList *core.TxList) error {
	out := make(chan error, len(*txList))
	for _, tx := range *txList {
		go func(tx *core.Transaction) {
			out <- pool.addNewTx(tx)
		}(tx)
	}
	for i := 0; i < len(*txList); i++ {
		err := <-out
		if err != nil {
			return err
		}
	}
	return nil
}

func (pool *TxPool) addNewTx(tx *core.Transaction) error {
	if err := tx.Validate(); err != nil {
		return err
	}
	if !pool.storage.HasTx(tx.Hash()) {
		pool.store.addNewTx(tx)
	}
	return nil
}

func (pool *TxPool) syncTxs(peer *core.PublicKey, hashes [][]byte) error {
	missing := make([][]byte, 0)
	for _, hash := range hashes {
		if !pool.storage.HasTx(hash) && pool.store.getTx(hash) == nil {
			missing = append(missing, hash)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	txList, err := pool.requestTxList(peer, missing)
	if err != nil {
		return err
	}
	return pool.addTxList(txList)
}

func (pool *TxPool) requestTxList(peer *core.PublicKey, hashes [][]byte) (*core.TxList, error) {
	txList, err := pool.msgSvc.RequestTxList(peer, hashes)
	if err != nil {
		return nil, err
	}
	for i, tx := range *txList {
		if !bytes.Equal(hashes[i], tx.Hash()) {
			return nil, errors.New("invalid txlist response")
		}
	}
	return txList, nil
}

func (pool *TxPool) verifyProposalTxs(hashes [][]byte) error {
	for _, hash := range hashes {
		if pool.storage.HasTx(hash) {
			return ErrOldTx
		}
		if pool.store.getTx(hash) == nil {
			return errors.New("tx not found")
		}
	}
	return nil
}

func (pool *TxPool) getTxsToExecute(hashes [][]byte) ([]*core.Transaction, [][]byte) {
	txs := make([]*core.Transaction, 0, len(hashes))
	executedTxs := make([][]byte, 0)
	for _, hash := range hashes {
		if pool.storage.HasTx(hash) {
			executedTxs = append(executedTxs, hash)
		} else {
			tx := pool.store.getTx(hash)
			if tx != nil {
				txs = append(txs, tx)
			} else {
				logger.I().Warnw("missing tx to execute")
			}
		}
	}
	pool.store.setTxsPending(hashes)
	return txs, executedTxs
}
