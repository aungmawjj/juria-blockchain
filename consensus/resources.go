// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package consensus

import (
	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/emitter"
	"github.com/aungmawjj/juria-blockchain/storage"
	"github.com/aungmawjj/juria-blockchain/txpool"
)

type TxPool interface {
	PopTxsFromQueue(max int) [][]byte
	SetTxsPending(hashes [][]byte)
	GetTxsToExecute(hashes [][]byte) ([]*core.Transaction, [][]byte)
	RemoveTxs(hashes [][]byte)
	PutTxsToQueue(hashes [][]byte)
	SyncTxs(peer *core.PublicKey, hashes [][]byte) error
	VerifyProposalTxs(hashes [][]byte) error
	GetStatus() txpool.Status
}

type Storage interface {
	GetMerkleRoot() []byte
	Commit(data *storage.CommitData) error
	GetBlock(hash []byte) (*core.Block, error)
	GetLastBlock() (*core.Block, error)
}

type MsgService interface {
	BroadcastProposal(blk *core.Block) error
	BroadcastNewView(qc *core.QuorumCert) error
	SendVote(pubKey *core.PublicKey, vote *core.Vote) error
	RequestBlock(pubKey *core.PublicKey, hash []byte) (*core.Block, error)
	SendNewView(pubKey *core.PublicKey, qc *core.QuorumCert) error

	SubscribeProposal(buffer int) *emitter.Subscription
	SubscribeVote(buffer int) *emitter.Subscription
	SubscribeNewView(buffer int) *emitter.Subscription
}

type Execution interface {
	Execute(blk *core.Block, txs []*core.Transaction) (*core.BlockCommit, []*core.TxCommit)
}

type Resources struct {
	Signer    core.Signer
	VldStore  core.ValidatorStore
	Storage   Storage
	MsgSvc    MsgService
	TxPool    TxPool
	Execution Execution
}
