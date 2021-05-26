// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package consensus

import (
	"bytes"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/hotstuff"
)

type blockStore interface {
	getBlock(hash []byte) *core.Block
}

type hsVote struct {
	vote  *core.Vote
	store blockStore
}

var _ hotstuff.Vote = (*hsVote)(nil)

func newHsVote(vote *core.Vote, store blockStore) hotstuff.Vote {
	return &hsVote{
		vote:  vote,
		store: store,
	}
}

func (v *hsVote) Block() hotstuff.Block {
	blk := v.store.getBlock(v.vote.BlockHash())
	if blk == nil {
		return nil
	}
	return newHsBlock(blk, v.store)
}

func (v *hsVote) Voter() string {
	voter := v.vote.Voter()
	if voter == nil {
		return ""
	}
	return voter.String()
}

type hsQC struct {
	qc    *core.QuorumCert
	store blockStore
}

func newHsQC(qc *core.QuorumCert, store blockStore) hotstuff.QC {
	return &hsQC{
		qc:    qc,
		store: store,
	}
}

func (q *hsQC) Block() hotstuff.Block {
	if q.qc == nil {
		return nil
	}
	blk := q.store.getBlock(q.qc.BlockHash())
	if blk == nil {
		return nil
	}
	return newHsBlock(blk, q.store)
}

type hsBlock struct {
	block *core.Block
	store blockStore
}

var _ hotstuff.Block = (*hsBlock)(nil)

func newHsBlock(block *core.Block, store blockStore) hotstuff.Block {
	return &hsBlock{
		block: block,
		store: store,
	}
}

func (b *hsBlock) Height() uint64 {
	return b.block.Height()
}

func (b *hsBlock) Parent() hotstuff.Block {
	blk := b.store.getBlock(b.block.ParentHash())
	if blk == nil {
		return nil
	}
	return newHsBlock(blk, b.store)
}

func (b *hsBlock) Equal(hsb hotstuff.Block) bool {
	if hsb == nil {
		return false
	}
	b2 := hsb.(*hsBlock)
	return bytes.Equal(b.block.Hash(), b2.block.Hash())
}

func (b *hsBlock) Justify() hotstuff.QC {
	return newHsQC(b.block.QuorumCert(), b.store)
}

func qcRefHeight(qc hotstuff.QC) (height uint64) {
	ref := qc.Block()
	if ref != nil {
		height = ref.Height()
	}
	return height
}
