// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package p2p

import (
	"errors"
	"time"

	p2p_pb "github.com/aungmawjj/juria-blockchain/p2p/pb"
	"github.com/aungmawjj/juria-blockchain/util/emitter"
	"google.golang.org/protobuf/proto"
)

type reqHandlerFunc func(request interface{}) (marshalable, error)

type reqHandler struct {
	unmarshalReq unmarshalFunc
	handler      reqHandlerFunc
}

func (hdlr *reqHandler) msgHandlerFunc(peer *Peer, msg *p2p_pb.Message) {
	go hdlr.handleRequest(peer, msg)
}

func (hdlr *reqHandler) handleRequest(peer *Peer, msg *p2p_pb.Message) {
	respMsg := new(p2p_pb.Message)
	respMsg.Seq = msg.Seq
	b, err := hdlr.invokeHandler(peer, msg)
	if err != nil {
		respMsg.Error = err.Error()
	}
	respMsg.Data = b
	respMsg.Type = p2p_pb.Message_Response
	msgB, _ := proto.Marshal(respMsg)
	peer.WriteMsg(msgB)
}

func (hdlr *reqHandler) invokeHandler(peer *Peer, msg *p2p_pb.Message) ([]byte, error) {
	req, err := hdlr.unmarshalReq(msg.Data)
	if err != nil {
		return nil, err
	}
	resp, err := hdlr.handler(req)
	if err != nil {
		return nil, err
	}
	return resp.Marshal()
}

type reqClient struct {
	peer            *Peer
	reqData         marshalable
	reqType         p2p_pb.Message_Type
	seq             uint32
	timeoutDuration time.Duration
	unmarshalResp   unmarshalFunc
}

func (client *reqClient) makeRequest() (interface{}, error) {
	sub := client.peer.SubscribeMsg()
	defer sub.Unsubscribe()

	client.sendRequest()
	return client.waitResponse(sub)
}

func (client *reqClient) sendRequest() error {
	b, err := client.reqData.Marshal()
	if err != nil {
		return err
	}

	msg := new(p2p_pb.Message)
	msg.Type = client.reqType
	msg.Data = b
	msg.Seq = client.seq

	msgB, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return client.peer.WriteMsg(msgB)
}

func (client *reqClient) waitResponse(sub *emitter.Subscription) (interface{}, error) {
	timeout := time.After(client.timeoutDuration)
	for {
		select {
		case <-timeout:
			return nil, errors.New("request timeout")

		case e := <-sub.Events():
			msgB := e.([]byte)
			respMsg := new(p2p_pb.Message)
			if err := proto.Unmarshal(msgB, respMsg); err != nil {
				break
			}
			if !(respMsg.Type == p2p_pb.Message_Response && respMsg.Seq == client.seq) {
				break
			}
			if len(respMsg.Error) > 0 {
				return nil, errors.New(respMsg.Error)
			}
			return client.unmarshalResp(respMsg.Data)
		}
	}
}
