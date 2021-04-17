// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package p2p

import (
	"errors"
	"time"

	"github.com/aungmawjj/juria-blockchain/p2p/p2p_pb"
	"github.com/aungmawjj/juria-blockchain/util/emitter"
	"google.golang.org/protobuf/proto"
)

type reqHandlerFunc func(request interface{}) (marshalable, error)

type reqHandler struct {
	unmarshalReq unmarshalFunc
	handler      reqHandlerFunc
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
	reqType         p2p_pb.Message_ReqType
	seq             uint32
	timeoutDuration time.Duration
	unmarshalResp   unmarshalFunc
}

func (client *reqClient) makeRequest() (interface{}, error) {
	if client.peer == nil {
		return nil, errors.New("peer not found")
	}

	sub := client.peer.SubscribeMsg()
	defer sub.Unsubscribe()
	respCh := client.getResponse(sub)

	client.sendRequest()
	return client.waitResponse(sub, respCh)
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
			resp := new(p2p_pb.Message)
			if err := proto.Unmarshal(msgB, resp); err != nil {
				continue
			}
			if !(resp.Type == p2p_pb.Message_Response && resp.Seq == client.seq) {
				continue
			}
			ch <- resp
		}
	}()
	return ch
}

func (client *reqClient) waitResponse(sub *emitter.Subscription, respCh <-chan *p2p_pb.Message) (interface{}, error) {
	timeout := time.After(client.timeoutDuration)
	for {
		select {
		case <-timeout:
			return nil, errors.New("request timeout")

		case resp := <-respCh:
			if len(resp.Error) > 0 {
				return nil, errors.New(resp.Error)
			}
			return client.unmarshalResp(resp.Data)
		}
	}
}
