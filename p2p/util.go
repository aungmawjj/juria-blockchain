// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package p2p

import (
	"reflect"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/p2p/p2p_pb"
	"google.golang.org/protobuf/proto"
)

var (
	unmarshalerBlock      unmarshalerFactory = castUnmarshalerFactory(core.NewBlock)
	unmarshalerVote                          = castUnmarshalerFactory(core.NewVote)
	unmarshalerTxList                        = castUnmarshalerFactory(core.NewTxList)
	unmarshalerQuorumCert                    = castUnmarshalerFactory(core.NewQuorumCert)
	unmarshalerHashList                      = castUnmarshalerFactory(core.NewHashList)
)

func castUnmarshalerFactory(unmarshalerFactory interface{}) unmarshalerFactory {
	return func() core.Unmarshaler {
		res := reflect.ValueOf(unmarshalerFactory).Call([]reflect.Value{})
		return res[0].Interface().(core.Unmarshaler)
	}
}

func castReqHandlerFunc(handler interface{}) reqHandlerFunc {
	return func(req interface{}) (core.Marshaler, error) {
		res := reflect.ValueOf(handler).Call([]reflect.Value{reflect.ValueOf(req)})
		err, _ := res[1].Interface().(error)
		return res[0].Interface().(core.Marshaler), err
	}
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
