// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package consensus

import "github.com/aungmawjj/juria-blockchain/core"

type TxPool interface {
	PopTxsFromQueue(max int) [][]byte
}

type Storage interface {
	GetMerkleRoot() []byte
}

type MsgService interface {
	BroadcastProposal(blk *core.Block) error
	SendVote(pubKey *core.PublicKey, vote *core.Vote) error
}
