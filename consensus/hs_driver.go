// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package consensus

import (
	"time"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/hotstuff"
	"github.com/aungmawjj/juria-blockchain/logger"
	"github.com/aungmawjj/juria-blockchain/storage"
)

type hsDriver struct {
	resources *Resources
	config    Config

	checkTxDelay time.Duration

	state *state
}

var _ hotstuff.Driver = (*hsDriver)(nil)

func (hsd *hsDriver) MajorityCount() int {
	return hsd.resources.VldStore.MajorityCount()
}

func (hsd *hsDriver) CreateLeaf(parent hotstuff.Block, qc hotstuff.QC, height uint64) hotstuff.Block {
	blk := core.NewBlock().
		SetParentHash(parent.(*hsBlock).block.Hash()).
		SetQuorumCert(qc.(*hsQC).qc).
		SetHeight(height).
		SetTransactions(hsd.resources.TxPool.PopTxsFromQueue(hsd.config.BlockTxLimit)).
		SetExecHeight(hsd.resources.Storage.GetBlockHeight()).
		SetMerkleRoot(hsd.resources.Storage.GetMerkleRoot()).
		SetTimestamp(time.Now().UnixNano()).
		Sign(hsd.resources.Signer)

	hsd.state.setBlock(blk)
	return newHsBlock(blk, hsd.state)
}

func (hsd *hsDriver) CreateQC(hsVotes []hotstuff.Vote) hotstuff.QC {
	votes := make([]*core.Vote, len(hsVotes))
	for i, hsv := range hsVotes {
		votes[i] = hsv.(*hsVote).vote
	}
	qc := core.NewQuorumCert().Build(votes)
	return newHsQC(qc, hsd.state)
}

func (hsd *hsDriver) BroadcastProposal(hsBlk hotstuff.Block) {
	blk := hsBlk.(*hsBlock).block
	hsd.resources.MsgSvc.BroadcastProposal(blk)
}

func (hsd *hsDriver) VoteBlock(hsBlk hotstuff.Block) {
	blk := hsBlk.(*hsBlock).block
	vote := blk.Vote(hsd.resources.Signer)
	hsd.resources.TxPool.SetTxsPending(blk.Transactions())
	hsd.delayVoteWhenNoTxs()
	proposer := hsd.resources.VldStore.GetValidatorIndex(blk.Proposer())
	if proposer != hsd.state.getLeaderIndex() {
		return // view changed happened
	}
	hsd.resources.MsgSvc.SendVote(blk.Proposer(), vote)
	logger.I().Debugw("voted block",
		"proposer", proposer,
		"height", hsBlk.Height(),
		"qc", qcRefHeight(hsBlk.Justify()),
	)
}

func (hsd *hsDriver) delayVoteWhenNoTxs() {
	timer := time.NewTimer(hsd.config.TxWaitTime)
	defer timer.Stop()
	for hsd.resources.TxPool.GetStatus().Total == 0 {
		select {
		case <-timer.C:
			return
		case <-time.After(hsd.checkTxDelay):
		}
	}
}

func (hsd *hsDriver) Commit(hsBlk hotstuff.Block) {
	bexe := hsBlk.(*hsBlock).block
	start := time.Now()
	txs, old := hsd.resources.TxPool.GetTxsToExecute(bexe.Transactions())
	bcm, txcs := hsd.resources.Execution.Execute(bexe, txs)
	bcm.SetOldBlockTxs(old)
	data := &storage.CommitData{
		Block:        bexe,
		QC:           hsd.state.getQC(bexe.Hash()),
		Transactions: txs,
		BlockCommit:  bcm,
		TxCommits:    txcs,
	}
	err := hsd.resources.Storage.Commit(data)
	if err != nil {
		logger.I().Fatalf("commit storage error: %+v", err)
	}
	hsd.state.addCommitedTxCount(len(txs))
	hsd.cleanStateOnCommited(bexe)
	logger.I().Debugw("commited bock",
		"height", bexe.Height(),
		"txs", len(txs),
		"elapsed", time.Since(start))
}

func (hsd *hsDriver) cleanStateOnCommited(bexec *core.Block) {
	hsd.resources.TxPool.RemoveTxs(bexec.Transactions())

	// qc for bexe is no longer needed here after commited to storage
	hsd.state.deleteQC(bexec.Hash())

	folks := hsd.state.getOlderBlocks(bexec.Height())
	for _, blk := range folks {
		// put transactions from folked block back to queue
		hsd.resources.TxPool.PutTxsToQueue(blk.Transactions())
		hsd.state.deleteQC(blk.Hash())
	}
	hsd.deleteMuchOlderBlocks(bexec)
}

func (hsd *hsDriver) deleteMuchOlderBlocks(bexec *core.Block) {
	height := int64(bexec.Height()) - 10
	if height < 0 {
		return
	}
	blks := hsd.state.getOlderBlocks(uint64(height))
	for _, blk := range blks {
		hsd.state.deleteBlock(blk.Hash())
	}
}
