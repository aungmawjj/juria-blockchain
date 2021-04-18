// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package p2p

import (
	"errors"
	"time"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/p2p/p2p_pb"
	"github.com/aungmawjj/juria-blockchain/util/emitter"
	"google.golang.org/protobuf/proto"
)

type reqMsgHandler struct {
	reqFactory unmarshalerFactory
	reqHandler reqHandler
}

func (hdlr *reqMsgHandler) handleReqMsg(peer *Peer, msg *p2p_pb.Message) {
	respMsg := new(p2p_pb.Message)
	respMsg.Seq = msg.Seq
	b, err := hdlr.invokeReqHandler(peer, msg)
	if err != nil {
		respMsg.Error = err.Error()
	}
	respMsg.Data = b
	respMsg.Type = p2p_pb.Message_Response
	msgB, _ := proto.Marshal(respMsg)
	peer.WriteMsg(msgB)
}

func (hdlr *reqMsgHandler) invokeReqHandler(peer *Peer, msg *p2p_pb.Message) ([]byte, error) {
	var req interface{} = msg.Data
	if hdlr.reqFactory != nil {
		reqObj := hdlr.reqFactory.Instance()
		if err := reqObj.Unmarshal(msg.Data); err != nil {
			return nil, err
		}
		req = reqObj
	}
	resp, err := hdlr.reqHandler.handleReq(req)
	if err != nil {
		return nil, err
	}
	return resp.Marshal()
}

type reqClient struct {
	peer            *Peer
	reqData         core.Marshaler
	reqType         p2p_pb.Message_ReqType
	seq             uint32
	timeoutDuration time.Duration
}

func (client *reqClient) makeRequest(resp core.Unmarshaler) error {
	if client.peer == nil {
		return errors.New("peer not found")
	}

	sub := client.peer.SubscribeMsg()
	defer sub.Unsubscribe()
	respCh := client.getResponse(sub)

	client.sendRequest()
	return client.waitResponse(sub, respCh, resp)
}

func (client *reqClient) sendRequest() error {
	b, err := client.reqData.Marshal()
	if err != nil {
		return err
	}

	msg := new(p2p_pb.Message)
	msg.Type = p2p_pb.Message_Request
	msg.Data = b
	msg.ReqType = client.reqType
	msg.Seq = client.seq

	msgB, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return client.peer.WriteMsg(msgB)
}

func (client *reqClient) getResponse(sub *emitter.Subscription) <-chan *p2p_pb.Message {
	ch := make(chan *p2p_pb.Message)
	go func() {
		for e := range sub.Events() {
			msgB := e.([]byte)
			respMsg := new(p2p_pb.Message)
			if err := proto.Unmarshal(msgB, respMsg); err != nil {
				continue
			}
			if !(respMsg.Type == p2p_pb.Message_Response && respMsg.Seq == client.seq) {
				continue
			}
			ch <- respMsg
		}
	}()
	return ch
}

func (client *reqClient) waitResponse(
	sub *emitter.Subscription, respCh <-chan *p2p_pb.Message, resp core.Unmarshaler,
) error {
	timeout := time.After(client.timeoutDuration)
	for {
		select {
		case <-timeout:
			return errors.New("request timeout")

		case respMsg := <-respCh:
			if len(respMsg.Error) > 0 {
				return errors.New(respMsg.Error)
			}
			return resp.Unmarshal(respMsg.Data)
		}
	}
}
