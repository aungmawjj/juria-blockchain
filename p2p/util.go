// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package p2p

import (
	"reflect"

	"github.com/aungmawjj/juria-blockchain/core"
	p2p_pb "github.com/aungmawjj/juria-blockchain/p2p/pb"
	"google.golang.org/protobuf/proto"
)

func castUnmarshalFunc(unmarshal interface{}) unmarshalFunc {
	return func(b []byte) (interface{}, error) {
		res := reflect.ValueOf(unmarshal).Call([]reflect.Value{reflect.ValueOf(b)})
		err, _ := res[1].Interface().(error)
		return res[0].Interface(), err
	}
}

var (
	unmarshalBlock         unmarshalFunc = castUnmarshalFunc(core.UnmarshalBlock)
	unmarshalVote                        = castUnmarshalFunc(core.UnmarshalVote)
	unmarshalQuorumCert                  = castUnmarshalFunc(core.UnmarshalQuorumCert)
	unmarshalTxList                      = castUnmarshalFunc(core.UnmarshalTxList)
	unmarshalHashList                    = castUnmarshalFunc(core.UnmarshalHashList)
	unmarshalBytesTypeCast               = castUnmarshalFunc(unmarshalBytesType)
)

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
