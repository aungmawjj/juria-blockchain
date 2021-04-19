// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package core

import (
	"github.com/aungmawjj/juria-blockchain/core/core_pb"
	"google.golang.org/protobuf/proto"
)

type StateChange struct {
	data *core_pb.StateChange
}

func NewStateChange() *StateChange {
	return &StateChange{
		data: new(core_pb.StateChange),
	}
}

func (sc *StateChange) Key() []byte           { return sc.data.Key }
func (sc *StateChange) Value() []byte         { return sc.data.Value }
func (sc *StateChange) PrevValue() []byte     { return sc.data.PrevValue }
func (sc *StateChange) TreeIndex() []byte     { return sc.data.TreeIndex }
func (sc *StateChange) PrevTreeIndex() []byte { return sc.data.PrevTreeIndex }

func (sc *StateChange) setData(val *core_pb.StateChange) *StateChange {
	sc.data = val
	return sc
}

func (sc *StateChange) SetKey(val []byte) *StateChange {
	sc.data.Key = val
	return sc
}

func (sc *StateChange) SetValue(val []byte) *StateChange {
	sc.data.Value = val
	return sc
}

func (sc *StateChange) SetPrevValue(val []byte) *StateChange {
	sc.data.PrevValue = val
	return sc
}

func (sc *StateChange) SetTreeIndex(val []byte) *StateChange {
	sc.data.TreeIndex = val
	return sc
}

func (sc *StateChange) SetPrevTreeIndex(val []byte) *StateChange {
	sc.data.PrevTreeIndex = val
	return sc
}

func (sc *StateChange) Marshal() ([]byte, error) {
	return proto.Marshal(sc.data)
}

func (sc *StateChange) Deleted() bool {
	return len(sc.Value()) == 0
}

func (sc *StateChange) Unmarshal(b []byte) error {
	data := new(core_pb.StateChange)
	if err := proto.Unmarshal(b, data); err != nil {
		return err
	}
	sc.setData(data)
	return nil
}
