// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package p2p

import (
	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/p2p/p2p_pb"
	"google.golang.org/protobuf/proto"
)

type BlockReqHandlerFunc func(hash []byte) (*core.Block, error)
type TxListReqHandlerFunc func(hashList *core.HashList) (*core.TxList, error)

type ReqHandlerFuncs struct {
	BlockReqHandler  BlockReqHandlerFunc
	TxListReqHandler TxListReqHandlerFunc
}

type msgHandlerFunc func(peer *Peer, msg *p2p_pb.Message)

type unmarshalerFactory interface {
	Instance() core.Unmarshaler
}

type blockFactory struct{}
type voteFactory struct{}
type quorumCertFactory struct{}
type txListFactory struct{}
type hashListFactory struct{}

func (f blockFactory) Instance() core.Unmarshaler      { return core.NewBlock() }
func (f voteFactory) Instance() core.Unmarshaler       { return core.NewVote() }
func (f quorumCertFactory) Instance() core.Unmarshaler { return core.NewQuorumCert() }
func (f txListFactory) Instance() core.Unmarshaler     { return core.NewTxList() }
func (f hashListFactory) Instance() core.Unmarshaler   { return core.NewHashList() }

type reqHandler interface {
	reqObjFactory() unmarshalerFactory
	handleReq(req interface{}) (core.Marshaler, error)
}

type blockReqHandler struct {
	fn BlockReqHandlerFunc
}

func (hdlr *blockReqHandler) reqObjFactory() unmarshalerFactory {
	return nil
}

func (hdlr *blockReqHandler) handleReq(req interface{}) (core.Marshaler, error) {
	return hdlr.fn(req.([]byte))
}

type txListReqHandler struct {
	fn TxListReqHandlerFunc
}

func (hdlr *txListReqHandler) reqObjFactory() unmarshalerFactory {
	return hashListFactory{}
}

func (hdlr *txListReqHandler) handleReq(req interface{}) (core.Marshaler, error) {
	return hdlr.fn(req.(*core.HashList))
}

func makeMsgBytes(msgType p2p_pb.Message_Type, data core.Marshaler) ([]byte, error) {
	b, err := data.Marshal()
	if err != nil {
		return nil, err
	}
	msg := new(p2p_pb.Message)
	msg.Type = msgType
	msg.Data = b

	return proto.Marshal(msg)
}

type bytesType []byte

func (b bytesType) Marshal() ([]byte, error) {
	return b, nil
}
