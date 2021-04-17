// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package p2p

import (
	"errors"
	"sync/atomic"
	"time"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/p2p/p2p_pb"
	"github.com/aungmawjj/juria-blockchain/util/emitter"
	"google.golang.org/protobuf/proto"
)

type msgHandlerFunc func(peer *Peer, msg *p2p_pb.Message)

type unmarshalerFactory func() core.Unmarshaler

type MsgService struct {
	host     *Host
	handlers map[p2p_pb.Message_Type]msgHandlerFunc

	proposalEmitter *emitter.Emitter
	voteEmitter     *emitter.Emitter
	newViewEmitter  *emitter.Emitter
	txListEmitter   *emitter.Emitter

	reqHandlers map[p2p_pb.Message_ReqType]*reqHandler
	reqSeq      uint32
}

type ReqHandlers struct {
	BlockReqHandler  func(hash []byte) (*core.Block, error)
	TxListReqHandler func(hashList *core.HashList) (*core.TxList, error)
}

func NewMsgService(host *Host, reqHandlers ReqHandlers) *MsgService {
	svc := new(MsgService)
	svc.host = host
	svc.host.SetPeerAddedHandler(svc.onAddedPeer)

	svc.proposalEmitter = emitter.New()
	svc.voteEmitter = emitter.New()
	svc.newViewEmitter = emitter.New()
	svc.txListEmitter = emitter.New()

	svc.setRequestHandlers(reqHandlers)
	svc.setMessageHandlers()
	return svc
}

func (svc *MsgService) setRequestHandlers(hdlrs ReqHandlers) {
	svc.reqHandlers = make(map[p2p_pb.Message_ReqType]*reqHandler)
	svc.reqHandlers[p2p_pb.Message_ReqBlock] = &reqHandler{
		handler: castReqHandlerFunc(hdlrs.BlockReqHandler),
	}
	svc.reqHandlers[p2p_pb.Message_ReqTxList] = &reqHandler{
		reqFactory: unmarshalerHashList,
		handler:    castReqHandlerFunc(hdlrs.TxListReqHandler),
	}
}

func (svc *MsgService) setMessageHandlers() {
	svc.handlers = make(map[p2p_pb.Message_Type]msgHandlerFunc)
	svc.handlers[p2p_pb.Message_Proposal] = svc.receiverHandlerFunc(unmarshalerBlock, svc.proposalEmitter)
	svc.handlers[p2p_pb.Message_Vote] = svc.receiverHandlerFunc(unmarshalerVote, svc.voteEmitter)
	svc.handlers[p2p_pb.Message_NewView] = svc.receiverHandlerFunc(unmarshalerQuorumCert, svc.newViewEmitter)
	svc.handlers[p2p_pb.Message_TxList] = svc.receiverHandlerFunc(unmarshalerTxList, svc.txListEmitter)
	svc.handlers[p2p_pb.Message_Request] = svc.handleMessageRequest
}

func (svc *MsgService) receiverHandlerFunc(unmarshalerFactory func() core.Unmarshaler, emitter *emitter.Emitter) msgHandlerFunc {
	return func(peer *Peer, msg *p2p_pb.Message) {
		obj := unmarshalerFactory()
		err := obj.Unmarshal(msg.Data)
		if err != nil {
			return
		}
		emitter.Emit(obj)
	}
}

func (svc *MsgService) handleMessageRequest(peer *Peer, msg *p2p_pb.Message) {
	if hdlr, ok := svc.reqHandlers[msg.ReqType]; ok {
		go hdlr.handleRequest(peer, msg)
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

func (svc *MsgService) BroadcastTxList(txList *core.TxList) error {
	return svc.broadcastData(p2p_pb.Message_TxList, txList)
}

func (svc *MsgService) broadcastData(msgType p2p_pb.Message_Type, data core.Marshaler) error {
	msgB, err := makeMsgBytes(msgType, data)
	if err != nil {
		return err
	}
	for _, peer := range svc.host.PeerStore().List() {
		peer.WriteMsg(msgB)
	}
	return nil
}

func (svc *MsgService) sendData(pubKey *core.PublicKey, msgType p2p_pb.Message_Type, data core.Marshaler) error {
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

func (svc *MsgService) RequestBlock(pubKey *core.PublicKey, hash []byte) (*core.Block, error) {
	seq := atomic.AddUint32(&svc.reqSeq, 1)
	client := &reqClient{
		peer:            svc.host.PeerStore().Load(pubKey),
		reqData:         bytesType(hash),
		reqType:         p2p_pb.Message_ReqBlock,
		seq:             seq,
		timeoutDuration: 2 * time.Second,
	}
	resp := core.NewBlock()
	if err := client.makeRequest(resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (svc *MsgService) RequestTxList(pubKey *core.PublicKey, hashList *core.HashList) (*core.TxList, error) {
	seq := atomic.AddUint32(&svc.reqSeq, 1)
	client := &reqClient{
		peer:            svc.host.PeerStore().Load(pubKey),
		reqData:         hashList,
		reqType:         p2p_pb.Message_ReqTxList,
		seq:             seq,
		timeoutDuration: 2 * time.Second,
	}
	resp := core.NewTxList()
	if err := client.makeRequest(resp); err != nil {
		return nil, err
	}
	return resp, nil
}
