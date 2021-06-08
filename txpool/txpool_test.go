// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package txpool

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/emitter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockStorage struct {
	mock.Mock
}

var _ Storage = (*MockStorage)(nil)

func (m *MockStorage) HasTx(hash []byte) bool {
	args := m.Called(hash)
	return args.Bool(0)
}

type MockExecution struct {
	mock.Mock
}

var _ Execution = (*MockExecution)(nil)

func (m *MockExecution) VerifyTx(tx *core.Transaction) error {
	args := m.Called(tx)
	return args.Error(0)
}

type MockMsgService struct {
	mock.Mock
}

var _ MsgService = (*MockMsgService)(nil)

func (m *MockMsgService) SubscribeTxList(buffer int) *emitter.Subscription {
	args := m.Called(buffer)
	return args.Get(0).(*emitter.Subscription)
}

func (m *MockMsgService) BroadcastTxList(txList *core.TxList) error {
	args := m.Called(txList)
	return args.Error(0)
}

func (m *MockMsgService) RequestTxList(pubKey *core.PublicKey, hashes [][]byte) (*core.TxList, error) {
	args := m.Called(pubKey, hashes)
	ret := args.Get(0)
	if ret == nil {
		return nil, args.Error(1)
	}
	return ret.(*core.TxList), args.Error(1)
}

func TestTxPool_SubmitTx(t *testing.T) {
	assert := assert.New(t)

	priv := core.GenerateKey(nil)

	storage := new(MockStorage)
	execution := new(MockExecution)
	msgSvc := new(MockMsgService)

	msgSvc.On("SubscribeTxList", mock.Anything).Return(emitter.New().Subscribe(10))

	pool := New(storage, execution, msgSvc)
	pool.broadcaster.timer.Reset(time.Hour) // to avoid timeout broadcast for testing
	pool.broadcaster.batchSize = 2          // broadcast after two successful submitTx

	time.Sleep(time.Millisecond)
	msgSvc.AssertExpectations(t)

	tx1 := core.NewTransaction().SetNonce(1).Sign(priv)
	tx2 := core.NewTransaction().SetNonce(2).Sign(priv)
	tx3 := core.NewTransaction().SetNonce(3).Sign(priv)

	storage.On("HasTx", tx1.Hash()).Return(false)
	execution.On("VerifyTx", tx1).Return(nil)
	err := pool.SubmitTx(tx1)

	assert.NoError(err)
	storage.AssertExpectations(t)
	execution.AssertExpectations(t)

	storage.On("HasTx", tx2.Hash()).Return(false)
	// tx2 is invalid to execute
	execution.On("VerifyTx", tx2).Return(fmt.Errorf("invalid tx"))
	err = pool.SubmitTx(tx2)

	assert.Error(err, "verify should failed for executed tx")
	storage.AssertExpectations(t)

	// tx3 is already executed
	storage.On("HasTx", tx3.Hash()).Return(true)
	msgSvc.On("BroadcastTxList", &core.TxList{tx1, tx3}).Return(nil)
	err = pool.SubmitTx(tx3)

	assert.NoError(err)
	storage.AssertExpectations(t)
	execution.AssertExpectations(t)

	time.Sleep(time.Millisecond)
	msgSvc.AssertExpectations(t)
	assert.Empty(pool.broadcaster.txBatch, "batch should be reset after broadcast")

	// only tx1 should be added to pool
	assert.Equal(1, pool.GetStatus().Queue)
}

func TestTxPool_SubscribeTxList(t *testing.T) {
	assert := assert.New(t)

	priv := core.GenerateKey(nil)

	storage := new(MockStorage)
	execution := new(MockExecution)
	msgSvc := new(MockMsgService)

	txEmitter := emitter.New()
	msgSvc.On("SubscribeTxList", mock.Anything).Return(txEmitter.Subscribe(10))

	pool := New(storage, execution, msgSvc)
	pool.broadcaster.timeout = time.Minute // to avoid timeout broadcast
	pool.broadcaster.timer.Reset(time.Minute)

	time.Sleep(time.Millisecond)

	tx1 := core.NewTransaction().SetNonce(1).Sign(priv)
	tx2 := core.NewTransaction().SetNonce(2).Sign(priv)
	tx3 := core.NewTransaction().SetNonce(3).Sign(priv)

	storage.On("HasTx", tx1.Hash()).Return(false)
	storage.On("HasTx", tx2.Hash()).Return(true)
	storage.On("HasTx", tx3.Hash()).Return(false)

	execution.On("VerifyTx", tx1).Return(nil)
	execution.On("VerifyTx", tx3).Return(nil)

	txEmitter.Emit(&core.TxList{tx1, tx2, tx3})

	time.Sleep(5 * time.Millisecond)

	assert.Equal(2, pool.GetStatus().Queue)
	storage.AssertExpectations(t)
}

func TestTxPool_Sync(t *testing.T) {
	assert := assert.New(t)

	priv := core.GenerateKey(nil)

	storage := new(MockStorage)
	execution := new(MockExecution)
	msgSvc := new(MockMsgService)

	msgSvc.On("SubscribeTxList", mock.Anything).Return(emitter.New().Subscribe(10))

	pool := New(storage, execution, msgSvc)
	pool.broadcaster.timeout = time.Minute // to avoid timeout broadcast
	pool.broadcaster.timer.Reset(time.Minute)

	time.Sleep(time.Millisecond)

	tx1 := core.NewTransaction().SetNonce(1).Sign(priv)
	tx2 := core.NewTransaction().SetNonce(2).Sign(priv)
	tx3 := core.NewTransaction().SetNonce(3).Sign(priv)

	storage.On("HasTx", tx1.Hash()).Return(false)
	storage.On("HasTx", tx2.Hash()).Return(true)
	storage.On("HasTx", tx3.Hash()).Return(false)

	execution.On("VerifyTx", tx1).Return(nil)
	execution.On("VerifyTx", tx3).Return(nil)

	err := pool.SyncTxs(priv.PublicKey(), [][]byte{tx2.Hash()})

	assert.NoError(err, "tx2 found in storage")

	pool.SubmitTx(tx1)

	msgSvc.On("RequestTxList", priv.PublicKey(), [][]byte{tx3.Hash()}).Once().
		Return(nil, errors.New("tx request failed"))
	err = pool.SyncTxs(priv.PublicKey(), [][]byte{tx1.Hash(), tx2.Hash(), tx3.Hash()})

	assert.Error(err, "tx request failed")
	msgSvc.AssertExpectations(t)

	msgSvc.On("RequestTxList", priv.PublicKey(), [][]byte{tx3.Hash()}).Once().
		Return(&core.TxList{tx2}, nil)
	err = pool.SyncTxs(priv.PublicKey(), [][]byte{tx1.Hash(), tx2.Hash(), tx3.Hash()})

	assert.Error(err, "wrong tx response")
	msgSvc.AssertExpectations(t)

	msgSvc.On("RequestTxList", priv.PublicKey(), [][]byte{tx3.Hash()}).Once().
		Return(&core.TxList{tx3}, nil)
	err = pool.SyncTxs(priv.PublicKey(), [][]byte{tx1.Hash(), tx2.Hash(), tx3.Hash()})

	assert.NoError(err)
	msgSvc.AssertExpectations(t)
	storage.AssertExpectations(t)
}

func TestTxPool_GetTxsToExecute(t *testing.T) {
	assert := assert.New(t)

	priv := core.GenerateKey(nil)

	storage := new(MockStorage)
	execution := new(MockExecution)
	msgSvc := new(MockMsgService)

	msgSvc.On("SubscribeTxList", mock.Anything).Return(emitter.New().Subscribe(10))

	pool := New(storage, execution, msgSvc)
	pool.broadcaster.timeout = time.Minute // to avoid timeout broadcast
	pool.broadcaster.timer.Reset(time.Minute)

	time.Sleep(time.Millisecond)

	tx1 := core.NewTransaction().SetNonce(1).Sign(priv)
	tx2 := core.NewTransaction().SetNonce(2).Sign(priv)
	tx3 := core.NewTransaction().SetNonce(3).Sign(priv)

	storage.On("HasTx", tx1.Hash()).Return(false)
	storage.On("HasTx", tx2.Hash()).Return(true)
	storage.On("HasTx", tx3.Hash()).Return(false)

	execution.On("VerifyTx", tx1).Return(nil)
	execution.On("VerifyTx", tx3).Return(nil)

	pool.SubmitTx(tx1)
	pool.SubmitTx(tx3)
	txs, old := pool.GetTxsToExecute([][]byte{tx1.Hash(), tx2.Hash(), tx3.Hash()})

	assert.Equal(2, len(txs))
	assert.Equal(tx1, txs[0])
	assert.Equal(tx3, txs[1])

	assert.Equal(1, len(old))
	assert.Equal(tx2.Hash(), old[0])
}
