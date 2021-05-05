package p2p

import (
	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/p2p/p2p_pb"
	"google.golang.org/protobuf/proto"
)

type ReqHandler interface {
	Type() p2p_pb.Request_Type
	HandleReq(sender *core.PublicKey, data []byte) ([]byte, error)
}

type TxListReqHandler struct {
	GetTxList func(sender *core.PublicKey, hashList [][]byte) (*core.TxList, error)
}

var _ ReqHandler = (*TxListReqHandler)(nil)

func (hdlr *TxListReqHandler) Type() p2p_pb.Request_Type {
	return p2p_pb.Request_TxList
}

func (hdlr *TxListReqHandler) HandleReq(sender *core.PublicKey, data []byte) ([]byte, error) {
	req := new(p2p_pb.HashList)
	if err := proto.Unmarshal(data, req); err != nil {
		return nil, err
	}
	txList, err := hdlr.GetTxList(sender, req.List)
	if err != nil {
		return nil, err
	}
	return txList.Marshal()
}

type BlockReqHandler struct {
	GetBlock func(sender *core.PublicKey, hash []byte) (*core.Block, error)
}

var _ ReqHandler = (*BlockReqHandler)(nil)

func (hdlr *BlockReqHandler) Type() p2p_pb.Request_Type {
	return p2p_pb.Request_Block
}

func (hdlr *BlockReqHandler) HandleReq(sender *core.PublicKey, data []byte) ([]byte, error) {
	block, err := hdlr.GetBlock(sender, data)
	if err != nil {
		return nil, err
	}
	return block.Marshal()
}
