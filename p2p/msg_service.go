// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package p2p

import (
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/emitter"
	"github.com/aungmawjj/juria-blockchain/p2p/p2p_pb"
	"google.golang.org/protobuf/proto"
)

type MsgType byte

const (
	_ MsgType = iota
	MsgTypeProposal
	MsgTypeVote
	MsgTypeNewView
	MsgTypeTxList
	MsgTypeRequest
	MsgTypeResponse
)

type msgReceiver func(peer *Peer, data []byte)

type MsgService struct {
	host      *Host
	receivers map[MsgType]msgReceiver

	proposalEmitter *emitter.Emitter
	voteEmitter     *emitter.Emitter
	newViewEmitter  *emitter.Emitter
	txListEmitter   *emitter.Emitter

	reqHandlers map[p2p_pb.Request_Type]ReqHandler

	reqClientSeq uint32
}

func NewMsgService(host *Host) *MsgService {
	svc := new(MsgService)
	svc.host = host
	svc.host.SetPeerAddedHandler(svc.onAddedPeer)
	for _, peer := range svc.host.PeerStore().List() {
		svc.host.onAddedPeer(peer)
	}

	svc.reqHandlers = make(map[p2p_pb.Request_Type]ReqHandler)
	svc.setEmitters()
	svc.setMsgReceivers()
	return svc
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
	data, err := blk.Marshal()
	if err != nil {
		return err
	}
	return svc.broadcastData(MsgTypeProposal, data)
}

func (svc *MsgService) SendVote(pubKey *core.PublicKey, vote *core.Vote) error {
	data, err := vote.Marshal()
	if err != nil {
		return err
	}
	return svc.sendData(pubKey, MsgTypeVote, data)
}

func (svc *MsgService) SendNewView(pubKey *core.PublicKey, qc *core.QuorumCert) error {
	data, err := qc.Marshal()
	if err != nil {
		return err
	}
	return svc.sendData(pubKey, MsgTypeNewView, data)
}

func (svc *MsgService) BroadcastNewView(qc *core.QuorumCert) error {
	data, err := qc.Marshal()
	if err != nil {
		return err
	}
	return svc.broadcastData(MsgTypeNewView, data)
}

func (svc *MsgService) BroadcastTxList(txList *core.TxList) error {
	data, err := txList.Marshal()
	if err != nil {
		return err
	}
	return svc.broadcastData(MsgTypeTxList, data)
}

func (svc *MsgService) RequestBlock(pubKey *core.PublicKey, hash []byte) (*core.Block, error) {
	respData, err := svc.requestData(pubKey, p2p_pb.Request_Block, hash)
	if err != nil {
		return nil, err
	}
	blk := core.NewBlock()
	return blk, blk.Unmarshal(respData)
}

func (svc *MsgService) RequestTxList(pubKey *core.PublicKey, hashes [][]byte) (*core.TxList, error) {
	hl := new(p2p_pb.HashList)
	hl.List = hashes
	reqData, _ := proto.Marshal(hl)
	respData, err := svc.requestData(pubKey, p2p_pb.Request_TxList, reqData)
	if err != nil {
		return nil, err
	}
	txList := core.NewTxList()
	return txList, txList.Unmarshal(respData)
}

func (svc *MsgService) SetReqHandler(reqHandler ReqHandler) error {
	if _, found := svc.reqHandlers[reqHandler.Type()]; found {
		return fmt.Errorf("request handler already set %s", reqHandler.Type())
	}
	svc.reqHandlers[reqHandler.Type()] = reqHandler
	return nil
}

func (svc *MsgService) setEmitters() {
	svc.proposalEmitter = emitter.New()
	svc.voteEmitter = emitter.New()
	svc.newViewEmitter = emitter.New()
	svc.txListEmitter = emitter.New()
}

func (svc *MsgService) setMsgReceivers() {
	svc.receivers = make(map[MsgType]msgReceiver)
	svc.receivers[MsgTypeProposal] = svc.onReceiveProposal
	svc.receivers[MsgTypeVote] = svc.onReceiveVote
	svc.receivers[MsgTypeNewView] = svc.onReceiveNewView
	svc.receivers[MsgTypeTxList] = svc.onReceiveTxList
	svc.receivers[MsgTypeRequest] = svc.onReceiveRequest
}

func (svc *MsgService) onAddedPeer(peer *Peer) {
	go svc.receivePeerMessages(peer)
}

func (svc *MsgService) receivePeerMessages(peer *Peer) {
	sub := peer.SubscribeMsg()
	for e := range sub.Events() {
		msg := e.([]byte)
		if len(msg) < 2 {
			continue // invalid message
		}
		if receiver, found := svc.receivers[MsgType(msg[0])]; found {
			receiver(peer, msg[1:])
		}
	}
}

func (svc *MsgService) onReceiveProposal(peer *Peer, data []byte) {
	blk := core.NewBlock()
	if err := blk.Unmarshal(data); err != nil {
		return
	}
	svc.proposalEmitter.Emit(blk)
}

func (svc *MsgService) onReceiveVote(peer *Peer, data []byte) {
	vote := core.NewVote()
	if err := vote.Unmarshal(data); err != nil {
		return
	}
	svc.voteEmitter.Emit(vote)
}

func (svc *MsgService) onReceiveNewView(peer *Peer, data []byte) {
	qc := core.NewQuorumCert()
	if err := qc.Unmarshal(data); err != nil {
		return
	}
	svc.newViewEmitter.Emit(qc)
}

func (svc *MsgService) onReceiveTxList(peer *Peer, data []byte) {
	txList := core.NewTxList()
	if err := txList.Unmarshal(data); err != nil {
		return
	}
	svc.txListEmitter.Emit(txList)
}

func (svc *MsgService) onReceiveRequest(peer *Peer, data []byte) {
	req := new(p2p_pb.Request)
	if err := proto.Unmarshal(data, req); err != nil {
		return
	}
	resp := new(p2p_pb.Response)
	resp.Seq = req.Seq

	if hdlr, found := svc.reqHandlers[req.Type]; found {
		data, err := hdlr.HandleReq(peer.PublicKey(), req.Data)
		if err != nil {
			resp.Error = err.Error()
		} else {
			resp.Data = data
		}
	} else {
		resp.Error = "no handler for request"
	}
	b, _ := proto.Marshal(resp)
	peer.WriteMsg(append([]byte{byte(MsgTypeResponse)}, b...))
}

func (svc *MsgService) broadcastData(msgType MsgType, data []byte) error {
	for _, peer := range svc.host.PeerStore().List() {
		peer.WriteMsg(append([]byte{byte(msgType)}, data...))
	}
	return nil
}

func (svc *MsgService) sendData(pubKey *core.PublicKey, msgType MsgType, data []byte) error {
	peer := svc.host.PeerStore().Load(pubKey)
	if peer == nil {
		return errors.New("peer not found")
	}
	return peer.WriteMsg(append([]byte{byte(msgType)}, data...))
}

func (svc *MsgService) requestData(
	pubKey *core.PublicKey, reqType p2p_pb.Request_Type, reqData []byte,
) ([]byte, error) {
	peer := svc.host.PeerStore().Load(pubKey)
	if peer == nil {
		return nil, errors.New("peer not found")
	}
	req := new(p2p_pb.Request)
	req.Type = reqType
	req.Data = reqData
	req.Seq = atomic.AddUint32(&svc.reqClientSeq, 1)
	b, _ := proto.Marshal(req)

	sub := peer.SubscribeMsg()
	defer sub.Unsubscribe()

	err := peer.WriteMsg(append([]byte{byte(MsgTypeRequest)}, b...))
	if err != nil {
		return nil, err
	}
	return svc.waitResponse(sub, req.Seq)
}

func (svc *MsgService) waitResponse(sub *emitter.Subscription, seq uint32) ([]byte, error) {
	timeout := time.After(5 * time.Second)
	for {
		select {
		case <-timeout:
			return nil, errors.New("request timeout")
		case e := <-sub.Events():
			msg := e.([]byte)
			if len(msg) < 2 {
				continue
			}
			if MsgType(msg[0]) == MsgTypeResponse {
				resp := new(p2p_pb.Response)
				if err := proto.Unmarshal(msg[1:], resp); err != nil {
					continue
				}
				if resp.Seq == seq {
					if len(resp.Error) > 0 {
						return nil, errors.New(resp.Error)
					}
					return resp.Data, nil
				}
			}
		}
	}
}
