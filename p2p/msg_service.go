// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package p2p

import (
	"errors"
	"time"

	"github.com/aungmawjj/juria-blockchain/core"
	p2p_pb "github.com/aungmawjj/juria-blockchain/p2p/pb"
	"github.com/aungmawjj/juria-blockchain/util/emitter"
	"google.golang.org/protobuf/proto"
)

type msgHandlerFunc func(peer *Peer, msg *p2p_pb.Message)

type unmarshalFunc func(b []byte) (interface{}, error)

type marshalable interface {
	Marshal() ([]byte, error)
}

type MsgService struct {
	host     *Host
	handlers map[p2p_pb.Message_Type]msgHandlerFunc

	proposalEmitter *emitter.Emitter
	voteEmitter     *emitter.Emitter
	newViewEmitter  *emitter.Emitter
	txListEmitter   *emitter.Emitter

	reqSeq uint32
}

func NewMsgService(host *Host) {
	svc := new(MsgService)
	svc.host = host
	svc.handlers = make(map[p2p_pb.Message_Type]msgHandlerFunc)
	svc.host.SetPeerAddedHandler(svc.onAddedPeer)

	svc.proposalEmitter = emitter.New()
	svc.voteEmitter = emitter.New()
	svc.newViewEmitter = emitter.New()
	svc.txListEmitter = emitter.New()

	svc.setReceiverHandlers()
}

func (svc *MsgService) setReceiverHandlers() {
	svc.handlers[p2p_pb.Message_Proposal] = svc.receiverHandlerFunc(core.UnmarshalBlock, svc.proposalEmitter)
	svc.handlers[p2p_pb.Message_Vote] = svc.receiverHandlerFunc(core.UnmarshalVote, svc.voteEmitter)
	svc.handlers[p2p_pb.Message_NewView] = svc.receiverHandlerFunc(core.UnmarshalQuorumCert, svc.newViewEmitter)
	svc.handlers[p2p_pb.Message_TxList] = svc.receiverHandlerFunc(core.UnmarshalTxList, svc.txListEmitter)
}

func (svc *MsgService) receiverHandlerFunc(unmarshal interface{}, emitter *emitter.Emitter) msgHandlerFunc {
	unmarshalData := unmarshal.(unmarshalFunc)
	return func(peer *Peer, msg *p2p_pb.Message) {
		data, err := unmarshalData(msg.Data)
		if err != nil {
			return
		}
		emitter.Emit(data)
	}
}

func (svc *MsgService) onAddedPeer(peer *Peer) {
	sub := peer.SubscribeMsg()
	for e := range sub.Events() {
		msgB := e.([]byte)
		msg := new(p2p_pb.Message)
		if err := proto.Unmarshal(msgB, msg); err != nil {
			continue
		}
		if handler, ok := svc.handlers[msg.Type]; ok {
			handler(peer, msg)
		}
	}
}

func (svc *MsgService) SetBlockReqHandler(handler func(hash []byte) (*core.Block, error)) {
	blkReqHandler := &reqHandler{
		unmarshalReq: (interface{})(unmarshalBytesType).(unmarshalFunc),
		handler:      (interface{})(handler).(reqHandlerFunc),
	}
	svc.handlers[p2p_pb.Message_BlockReq] = blkReqHandler.msgHandlerFunc
}

func (svc *MsgService) SetTxListReqHandler(handler func(hashList core.HashList) (core.TxList, error)) {
	txListReqHandler := &reqHandler{
		unmarshalReq: (interface{})(core.UnmarshalTxList).(unmarshalFunc),
		handler:      (interface{})(handler).(reqHandlerFunc),
	}
	svc.handlers[p2p_pb.Message_TxListReq] = txListReqHandler.msgHandlerFunc
}

func (svc *MsgService) RequestBlock(pubKey *core.PublicKey, hash []byte) (*core.Block, error) {
	peer := svc.host.PeerStore().Load(pubKey)
	if peer != nil {
		return nil, errors.New("peer not found")
	}

	svc.reqSeq++
	client := &reqClient{
		peer:            peer,
		reqData:         bytesType(hash),
		reqType:         p2p_pb.Message_BlockReq,
		seq:             svc.reqSeq,
		timeoutDuration: 2 * time.Second,
		unmarshalResp:   (interface{})(core.UnmarshalBlock).(unmarshalFunc),
	}
	resp, err := client.makeRequest()
	if err != nil {
		return nil, err
	}
	return (resp).(*core.Block), nil
}

func (svc *MsgService) RequestTxList(pubKey *core.PublicKey, hashList core.HashList) (*core.TxList, error) {
	peer := svc.host.PeerStore().Load(pubKey)
	if peer != nil {
		return nil, errors.New("peer not found")
	}

	svc.reqSeq++
	client := &reqClient{
		peer:            peer,
		reqData:         hashList,
		reqType:         p2p_pb.Message_TxListReq,
		seq:             svc.reqSeq,
		timeoutDuration: 4 * time.Second,
		unmarshalResp:   (interface{})(core.UnmarshalTxList).(unmarshalFunc),
	}
	resp, err := client.makeRequest()
	if err != nil {
		return nil, err
	}
	return (resp).(*core.TxList), nil
}

type bytesType []byte

func (b bytesType) Marshal() ([]byte, error) {
	return b, nil
}

func unmarshalBytesType(b []byte) ([]byte, error) {
	return b, nil
}
