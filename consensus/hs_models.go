// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package consensus

import (
	"bytes"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/hotstuff"
)

type blockLoader interface {
	loadBlock(hash []byte) (*core.Block, bool)
}

type hsVote struct {
	vote        *core.Vote
	blockLoader blockLoader
}

var _ hotstuff.Vote = (*hsVote)(nil)

func newHsVote(vote *core.Vote, blockLoader blockLoader) hotstuff.Vote {
	return &hsVote{
		vote:        vote,
		blockLoader: blockLoader,
	}
}

func (v *hsVote) Block() hotstuff.Block {
	blk, ok := v.blockLoader.loadBlock(v.vote.BlockHash())
	if !ok {
		return nil
	}
	return newHsBlock(blk, v.blockLoader)
}

func (v *hsVote) Voter() string {
	voter := v.vote.Voter()
	if voter == nil {
		return ""
	}
	return voter.String()
}

type hsQC struct {
	qc          *core.QuorumCert
	blockLoader blockLoader
}

func newHsQC(qc *core.QuorumCert, blockLoader blockLoader) hotstuff.QC {
	return &hsQC{
		qc:          qc,
		blockLoader: blockLoader,
	}
}

func (q *hsQC) Block() hotstuff.Block {
	blk, ok := q.blockLoader.loadBlock(q.qc.BlockHash())
	if !ok {
		return nil
	}
	return newHsBlock(blk, q.blockLoader)
}

type hsBlock struct {
	block       *core.Block
	blockLoader blockLoader
}

var _ hotstuff.Block = (*hsBlock)(nil)

func newHsBlock(block *core.Block, blockLoader blockLoader) hotstuff.Block {
	return &hsBlock{
		block:       block,
		blockLoader: blockLoader,
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
	blk, ok := b.blockLoader.loadBlock(b.block.ParentHash())
	if !ok {
		return nil
	}
	return newHsBlock(blk, b.blockLoader)
}

func (b *hsBlock) Equal(hsb hotstuff.Block) bool {
	if hsb == nil {
		return false
	}
	b2 := hsb.(*hsBlock)
	return bytes.Equal(b.block.Hash(), b2.block.Hash())
}

func (b *hsBlock) Justify() hotstuff.QC {
	return newHsQC(b.block.QuorumCert(), b.blockLoader)
}
