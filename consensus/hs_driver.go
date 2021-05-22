// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package consensus

import (
	"context"
	"time"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/hotstuff"
)

type hsDriver struct {
	txPool   TxPool
	storage  Storage
	msgSvc   MsgService
	hotstuff *hotstuff.Hotstuff
	vstore   core.ValidatorStore
	signer   core.Signer
	store    *blockStore

	maxTxCount int
	txWaitTime time.Duration
}

var _ hotstuff.Driver = (*hsDriver)(nil)

func (hsd *hsDriver) CreateLeaf(
	ctx context.Context, parent hotstuff.Block, qc hotstuff.QC, height uint64,
) hotstuff.Block {

	blk := core.NewBlock().
		SetParentHash(parent.(*hsBlock).block.ParentHash()).
		SetQuorumCert(qc.(*hsQC).qc).
		SetHeight(height)

	return hsd.buildBlockWithTxs(ctx, blk)
}

func (hsd *hsDriver) buildBlockWithTxs(ctx context.Context, blk *core.Block) hotstuff.Block {
	timer := time.NewTimer(hsd.txWaitTime)
	for {
		select {
		case <-ctx.Done(): // canceled
			return nil

		case <-time.After(time.Millisecond):
			if txs := hsd.txPool.PopTxsFromQueue(hsd.maxTxCount); txs != nil {
				blk.SetTransactions(txs)
				hsd.sealAndStoreBlock(blk)
				return newHsBlock(blk, hsd.store)
			}

		case <-timer.C: // create empty block
			hsd.sealAndStoreBlock(blk)
			return newHsBlock(blk, hsd.store)
		}
	}
}

func (hsd *hsDriver) sealAndStoreBlock(blk *core.Block) {
	blk.SetExecHeight(hsd.hotstuff.GetBExec().Height()).
		SetMerkleRoot(hsd.storage.GetMerkleRoot()).
		SetTimestamp(time.Now().UnixNano()).
		Sign(hsd.signer)
	hsd.store.setBlock(blk)
}

func (hsd *hsDriver) CreateQC(hsVotes []hotstuff.Vote) hotstuff.QC {
	votes := make([]*core.Vote, len(hsVotes))
	for i, hsv := range hsVotes {
		votes[i] = hsv.(*hsVote).vote
	}
	qc := core.NewQuorumCert().Build(votes)
	return newHsQC(qc, hsd.store)
}

func (hsd *hsDriver) BroadcastProposal(hsBlk hotstuff.Block) {
	blk := hsBlk.(*hsBlock).block
	hsd.msgSvc.BroadcastProposal(blk)
}

func (hsd *hsDriver) VoteBlock(hsBlk hotstuff.Block) {
	blk := hsBlk.(*hsBlock).block
	vote := blk.Vote(hsd.signer)
	if hsd.signer.PublicKey().Equal(blk.Proposer()) {
		hsd.hotstuff.OnReceiveVote(newHsVote(vote, hsd.store))
	} else {
		hsd.msgSvc.SendVote(blk.Proposer(), vote)
	}
}

func (hsd *hsDriver) Execute(hsBlk hotstuff.Block) {

}

func (hsd *hsDriver) MajorityCount() int {
	return hsd.vstore.MajorityCount()
}
