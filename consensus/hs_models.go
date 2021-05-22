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
	vote       *core.Vote
	blockStore blockStore
}

var _ hotstuff.Vote = (*hsVote)(nil)

func newHsVote(vote *core.Vote, blockLoader blockStore) hotstuff.Vote {
	return &hsVote{
		vote:       vote,
		blockStore: blockLoader,
	}
}

func (v *hsVote) Block() hotstuff.Block {
	blk := v.blockStore.getBlock(v.vote.BlockHash())
	if blk == nil {
		return nil
	}
	return newHsBlock(blk, v.blockStore)
}

func (v *hsVote) Voter() string {
	voter := v.vote.Voter()
	if voter == nil {
		return ""
	}
	return voter.String()
}

type hsQC struct {
	qc         *core.QuorumCert
	blockStore blockStore
}

func newHsQC(qc *core.QuorumCert, blockLoader blockStore) hotstuff.QC {
	return &hsQC{
		qc:         qc,
		blockStore: blockLoader,
	}
}

func (q *hsQC) Block() hotstuff.Block {
	blk := q.blockStore.getBlock(q.qc.BlockHash())
	if blk == nil {
		return nil
	}
	return newHsBlock(blk, q.blockStore)
}

type hsBlock struct {
	block      *core.Block
	blockStore blockStore
}

var _ hotstuff.Block = (*hsBlock)(nil)

func newHsBlock(block *core.Block, blockLoader blockStore) hotstuff.Block {
	return &hsBlock{
		block:      block,
		blockStore: blockLoader,
	}
}

func (b *hsBlock) Proposer() string {
	p := b.block.Proposer()
	if p == nil {
		return ""
	}
	return p.String()
}

func (b *hsBlock) Height() uint64 {
	return b.block.Height()
}

func (b *hsBlock) Parent() hotstuff.Block {
	blk := b.blockStore.getBlock(b.block.ParentHash())
	if blk == nil {
		return nil
	}
	return newHsBlock(blk, b.blockStore)
}

func (b *hsBlock) Equal(hsb hotstuff.Block) bool {
	if hsb == nil {
		return false
	}
	b2 := hsb.(*hsBlock)
	return bytes.Equal(b.block.Hash(), b2.block.Hash())
}

func (b *hsBlock) Justify() hotstuff.QC {
	return newHsQC(b.block.QuorumCert(), b.blockStore)
}
