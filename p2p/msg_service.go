// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package p2p

import (
	"errors"
	"sync/atomic"
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

func NewMsgService(host *Host) *MsgService {
	svc := new(MsgService)
	svc.host = host
	svc.handlers = make(map[p2p_pb.Message_Type]msgHandlerFunc)
	svc.host.SetPeerAddedHandler(svc.onAddedPeer)

	svc.proposalEmitter = emitter.New()
	svc.voteEmitter = emitter.New()
	svc.newViewEmitter = emitter.New()
	svc.txListEmitter = emitter.New()

	svc.setReceiverHandlers()
	return svc
}

func (svc *MsgService) setReceiverHandlers() {
	svc.handlers[p2p_pb.Message_Proposal] = svc.receiverHandlerFunc(unmarshalBlock, svc.proposalEmitter)
	svc.handlers[p2p_pb.Message_Vote] = svc.receiverHandlerFunc(unmarshalVote, svc.voteEmitter)
	svc.handlers[p2p_pb.Message_NewView] = svc.receiverHandlerFunc(unmarshalQuorumCert, svc.newViewEmitter)
	svc.handlers[p2p_pb.Message_TxList] = svc.receiverHandlerFunc(unmarshalTxList, svc.txListEmitter)
}

func (svc *MsgService) receiverHandlerFunc(unmarshal unmarshalFunc, emitter *emitter.Emitter) msgHandlerFunc {
	return func(peer *Peer, msg *p2p_pb.Message) {
		data, err := unmarshal(msg.Data)
		if err != nil {
			return
		}
		emitter.Emit(data)
	}
}

func (svc *MsgService) onAddedPeer(peer *Peer) {
	go svc.handlePeerMsg(peer)
}

func (svc *MsgService) handlePeerMsg(peer *Peer) {
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

func (svc *MsgService) SubscribeProposal(buffer int) *emitter.Subscription {
	return svc.proposalEmitter.Subscribe(buffer)
}

func (svc *MsgService) SubscribeVote(buffer int) *emitter.Subscription {
	return svc.voteEmitter.Subscribe(buffer)
}

func (svc *MsgService) SubscribeNewView(buffer int) *emitter.Subscription {
	return svc.newViewEmitter.Subscribe(buffer)
}

func (svc *MsgService) SubscribeTxList(buffer int) *emitter.Subscription {
	return svc.txListEmitter.Subscribe(buffer)
}

func (svc *MsgService) BroadcastProposal(blk *core.Block) error {
	return svc.broadcastData(p2p_pb.Message_Proposal, blk)
}

func (svc *MsgService) SendVote(pubKey *core.PublicKey, vote *core.Vote) error {
	return svc.sendData(pubKey, p2p_pb.Message_Vote, vote)
}

func (svc *MsgService) SendNewView(pubKey *core.PublicKey, qc *core.QuorumCert) error {
	return svc.sendData(pubKey, p2p_pb.Message_NewView, qc)
}

func (svc *MsgService) BroadcastTxList(txList core.TxList) error {
	return svc.broadcastData(p2p_pb.Message_TxList, txList)
}

func (svc *MsgService) broadcastData(msgType p2p_pb.Message_Type, data marshalable) error {
	msgB, err := makeMsgBytes(msgType, data)
	if err != nil {
		return err
	}
	for _, peer := range svc.host.PeerStore().List() {
		peer.WriteMsg(msgB)
	}
	return nil
}

func (svc *MsgService) sendData(pubKey *core.PublicKey, msgType p2p_pb.Message_Type, data marshalable) error {
	peer := svc.host.PeerStore().Load(pubKey)
	if peer == nil {
		return errors.New("peer not found")
	}
	msgB, err := makeMsgBytes(msgType, data)
	if err != nil {
		return err
	}
	return peer.WriteMsg(msgB)
}

func (svc *MsgService) SetBlockReqHandler(handler func(hash []byte) (*core.Block, error)) {
	blkReqHandler := &reqHandler{
		unmarshalReq: unmarshalBytesTypeCast,
		handler:      (interface{})(handler).(reqHandlerFunc),
	}
	svc.handlers[p2p_pb.Message_BlockReq] = blkReqHandler.msgHandlerFunc
}

func (svc *MsgService) SetTxListReqHandler(handler func(hashList core.HashList) (core.TxList, error)) {
	txListReqHandler := &reqHandler{
		unmarshalReq: unmarshalHashList,
		handler:      (interface{})(handler).(reqHandlerFunc),
	}
	svc.handlers[p2p_pb.Message_TxListReq] = txListReqHandler.msgHandlerFunc
}

func (svc *MsgService) RequestBlock(pubKey *core.PublicKey, hash []byte) (*core.Block, error) {
	seq := atomic.AddUint32(&svc.reqSeq, 1)
	client := &reqClient{
		peer:            svc.host.PeerStore().Load(pubKey),
		reqData:         bytesType(hash),
		reqType:         p2p_pb.Message_BlockReq,
		seq:             seq,
		timeoutDuration: 2 * time.Second,
		unmarshalResp:   unmarshalBlock,
	}
	resp, err := client.makeRequest()
	if err != nil {
		return nil, err
	}
	return (resp).(*core.Block), nil
}

func (svc *MsgService) RequestTxList(pubKey *core.PublicKey, hashList core.HashList) (*core.TxList, error) {
	seq := atomic.AddUint32(&svc.reqSeq, 1)
	client := &reqClient{
		peer:            svc.host.PeerStore().Load(pubKey),
		reqData:         hashList,
		reqType:         p2p_pb.Message_TxListReq,
		seq:             seq,
		timeoutDuration: 2 * time.Second,
		unmarshalResp:   unmarshalTxList,
	}
	resp, err := client.makeRequest()
	if err != nil {
		return nil, err
	}
	return (resp).(*core.TxList), nil
}
