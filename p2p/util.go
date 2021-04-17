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
	unmarshalBlock         unmarshalFunc = castUnmarshalFunc(core.UnmarshalBlock)
	unmarshalVote                        = castUnmarshalFunc(core.UnmarshalVote)
	unmarshalQuorumCert                  = castUnmarshalFunc(core.UnmarshalQuorumCert)
	unmarshalTxList                      = castUnmarshalFunc(core.UnmarshalTxList)
	unmarshalHashList                    = castUnmarshalFunc(core.UnmarshalHashList)
	unmarshalBytesTypeCast               = castUnmarshalFunc(unmarshalBytesType)
)

func castUnmarshalFunc(unmarshal interface{}) unmarshalFunc {
	return func(b []byte) (interface{}, error) {
		res := reflect.ValueOf(unmarshal).Call([]reflect.Value{reflect.ValueOf(b)})
		err, _ := res[1].Interface().(error)
		return res[0].Interface(), err
	}
}

func castReqHandlerFunc(handler interface{}) reqHandlerFunc {
	return func(req interface{}) (marshalable, error) {
		res := reflect.ValueOf(handler).Call([]reflect.Value{reflect.ValueOf(req)})
		err, _ := res[1].Interface().(error)
		return res[0].Interface().(marshalable), err
	}
}

func makeMsgBytes(msgType p2p_pb.Message_Type, data marshalable) ([]byte, error) {
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

func unmarshalBytesType(b []byte) ([]byte, error) {
	return b, nil
}
