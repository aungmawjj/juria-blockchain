// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package consensus

import (
	"testing"
	"time"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/hotstuff"
	"github.com/aungmawjj/juria-blockchain/storage"
	"github.com/aungmawjj/juria-blockchain/txpool"
	"github.com/stretchr/testify/assert"
)

func setupTestHsDriver() *hsDriver {
	resources := &Resources{
		Signer: core.GenerateKey(nil),
	}
	state := newState(resources)
	return &hsDriver{
		resources: resources,
		config:    DefaultConfig,
		state:     state,
	}
}

func TestHsDriver_TestMajorityCount(t *testing.T) {
	hsd := setupTestHsDriver()
	hsd.resources.VldStore = core.NewValidatorStore([]*core.PublicKey{
		core.GenerateKey(nil).PublicKey(),
		core.GenerateKey(nil).PublicKey(),
		core.GenerateKey(nil).PublicKey(),
		core.GenerateKey(nil).PublicKey(),
	})

	res := hsd.MajorityCount()

	assert := assert.New(t)
	assert.Equal(hsd.resources.VldStore.MajorityCount(), res)
}

func TestHsDriver_CreateLeaf(t *testing.T) {
	hsd := setupTestHsDriver()
	parent := newHsBlock(core.NewBlock().Sign(hsd.resources.Signer), hsd.state)
	hsd.state.setBlock(parent.(*hsBlock).block)
	qc := newHsQC(core.NewQuorumCert(), hsd.state)
	height := uint64(5)

	txsInQ := [][]byte{[]byte("tx1"), []byte("tx2")}
	txPool := new(MockTxPool)
	txPool.On("PopTxsFromQueue", hsd.config.BlockTxLimit).Return(txsInQ)
	hsd.resources.TxPool = txPool

	storage := new(MockStorage)
	storage.On("GetBlockHeight").Return(2) // driver should get bexec height from storage
	storage.On("GetMerkleRoot").Return([]byte("merkle-root"))
	hsd.resources.Storage = storage

	leaf := hsd.CreateLeaf(parent, qc, height)

	txPool.AssertExpectations(t)
	storage.AssertExpectations(t)

	assert := assert.New(t)
	assert.NotNil(leaf)
	assert.True(parent.Equal(leaf.Parent()), "should link to parent")
	assert.Equal(qc, leaf.Justify(), "should add qc")
	assert.Equal(height, leaf.Height())

	blk := leaf.(*hsBlock).block
	assert.Equal(txsInQ, blk.Transactions())
	assert.EqualValues(2, blk.ExecHeight())
	assert.Equal([]byte("merkle-root"), blk.MerkleRoot())
	assert.NotEmpty(blk.Timestamp(), "should add timestamp")

	assert.NotNil(hsd.state.getBlock(blk.Hash()), "should store leaf block in state")
}

func TestHsDriver_VoteBlock(t *testing.T) {
	hsd := setupTestHsDriver()
	hsd.checkTxDelay = time.Millisecond
	hsd.config.TxWaitTime = 2 * time.Millisecond

	proposer := core.GenerateKey(nil)
	blk := core.NewBlock().Sign(proposer)

	hsd.resources.VldStore = core.NewValidatorStore([]*core.PublicKey{blk.Proposer()})

	txPool := new(MockTxPool)
	txPool.On("GetStatus").Return(txpool.Status{}) // no txs in the pool
	txPool.On("SetTxsPending", blk.Transactions())
	hsd.resources.TxPool = txPool

	// should sign block and send vote
	msgSvc := new(MockMsgService)
	msgSvc.On("SendVote", proposer.PublicKey(), blk.Vote(hsd.resources.Signer)).Return(nil)
	hsd.resources.MsgSvc = msgSvc

	start := time.Now()
	hsd.VoteBlock(newHsBlock(blk, hsd.state))
	elapsed := time.Since(start)

	txPool.AssertExpectations(t)
	msgSvc.AssertExpectations(t)

	assert := assert.New(t)
	assert.GreaterOrEqual(elapsed, hsd.config.TxWaitTime, "should delay if no txs in the pool")

	txPool = new(MockTxPool)
	hsd.resources.TxPool = txPool
	txPool.On("GetStatus").Return(txpool.Status{Total: 1}) // one txs in the pool
	txPool.On("SetTxsPending", blk.Transactions())

	start = time.Now()
	hsd.VoteBlock(newHsBlock(blk, hsd.state))
	elapsed = time.Since(start)

	txPool.AssertExpectations(t)
	msgSvc.AssertExpectations(t)

	assert.Less(elapsed, hsd.config.TxWaitTime, "should not delay if txs in the pool")
}

func TestHsDriver_Commit(t *testing.T) {
	hsd := setupTestHsDriver()
	parent := core.NewBlock().SetHeight(10).Sign(hsd.resources.Signer)
	bfolk := core.NewBlock().SetTransactions([][]byte{[]byte("txfromfolk")}).SetHeight(10).Sign(hsd.resources.Signer)

	tx := core.NewTransaction().Sign(hsd.resources.Signer)
	bexec := core.NewBlock().SetTransactions([][]byte{tx.Hash()}).
		SetParentHash(parent.Hash()).SetHeight(11).Sign(hsd.resources.Signer)
	hsd.state.setBlock(parent)
	hsd.state.setCommitedBlock(parent)
	hsd.state.setBlock(bfolk)
	hsd.state.setBlock(bexec)

	txs := []*core.Transaction{tx}
	txPool := new(MockTxPool)
	txPool.On("GetTxsToExecute", bexec.Transactions()).Return(txs, nil)
	// should remove txs from pool after commit
	txPool.On("RemoveTxs", bexec.Transactions()).Once()
	// should put txs of folked block back to queue from pending
	txPool.On("PutTxsToQueue", bfolk.Transactions()).Once()
	hsd.resources.TxPool = txPool

	bcm := core.NewBlockCommit().SetHash(bexec.Hash())
	txcs := []*core.TxCommit{core.NewTxCommit().SetHash(tx.Hash())}
	execution := new(MockExecution)
	execution.On("Execute", bexec, txs).Return(bcm, txcs)
	hsd.resources.Execution = execution

	cdata := &storage.CommitData{
		Block:        bexec,
		Transactions: txs,
		BlockCommit:  bcm,
		TxCommits:    txcs,
	}
	storage := new(MockStorage)
	storage.On("Commit", cdata).Return(nil)
	hsd.resources.Storage = storage

	hsd.Commit(newHsBlock(bexec, hsd.state))

	txPool.AssertExpectations(t)
	execution.AssertExpectations(t)
	storage.AssertExpectations(t)

	assert := assert.New(t)
	assert.NotNil(hsd.state.getBlockFromState(bexec.Hash()),
		"should not delete bexec from state")
	assert.Nil(hsd.state.getBlockFromState(bfolk.Hash()),
		"should delete folked block from state")
}

func TestHsDriver_CreateQC(t *testing.T) {
	hsd := setupTestHsDriver()
	blk := core.NewBlock().Sign(hsd.resources.Signer)
	hsd.state.setBlock(blk)
	votes := []hotstuff.Vote{
		newHsVote(blk.ProposerVote(), hsd.state),
		newHsVote(blk.Vote(core.GenerateKey(nil)), hsd.state),
	}
	qc := hsd.CreateQC(votes)

	assert := assert.New(t)
	assert.Equal(blk, qc.Block().(*hsBlock).block, "should get qc reference block")
}

func TestHsDriver_BroadcastProposal(t *testing.T) {
	hsd := setupTestHsDriver()
	blk := core.NewBlock().Sign(hsd.resources.Signer)
	hsd.state.setBlock(blk)

	msgSvc := new(MockMsgService)
	msgSvc.On("BroadcastProposal", blk).Return(nil)
	hsd.resources.MsgSvc = msgSvc

	hsd.BroadcastProposal(newHsBlock(blk, hsd.state))

	msgSvc.AssertExpectations(t)
}
