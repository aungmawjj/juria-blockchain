// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package chaincode

type MockState struct {
	StateMap map[string][]byte
}

func NewMockState() *MockState {
	return &MockState{
		StateMap: make(map[string][]byte),
	}
}

func (ms *MockState) GetState(key []byte) []byte {
	return ms.StateMap[string(key)]
}

func (ms *MockState) SetState(key, value []byte) {
	ms.StateMap[string(key)] = value
}

type MockCallContext struct {
	MockSender      []byte
	MockBlockHeight uint64
	MockBlockHash   []byte
	MockInput       []byte
}

func (wc *MockCallContext) Sender() []byte {
	return wc.MockSender
}

func (wc *MockCallContext) BlockHash() []byte {
	return wc.MockBlockHash
}

func (wc *MockCallContext) BlockHeight() uint64 {
	return wc.MockBlockHeight
}

func (wc *MockCallContext) Input() []byte {
	return wc.MockInput
}

type MockReadContext struct {
	MockCallContext
	GetStateError error
	State         *MockState
}

var _ ReadContext = (*MockReadContext)(nil)

func (rc *MockReadContext) Input() []byte {
	return rc.MockInput
}

func (rc *MockReadContext) GetState(key []byte) ([]byte, error) {
	if rc.GetStateError != nil {
		return nil, rc.GetStateError
	}
	return rc.State.GetState(key), nil
}

type MockWriteContext struct {
	MockCallContext
	State *MockState
}

func (wc *MockWriteContext) GetState(key []byte) []byte {
	return wc.State.GetState(key)
}

func (wc *MockWriteContext) SetState(key, value []byte) {
	wc.State.SetState(key, value)
}
