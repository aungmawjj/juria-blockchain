// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package chaincode

type MockState struct {
	StateMap    map[string][]byte
	VerifyError error
}

func NewMockState() *MockState {
	return &MockState{
		StateMap: make(map[string][]byte),
	}
}

func (ms *MockState) VerifyState(key []byte) ([]byte, error) {
	if ms.VerifyError != nil {
		return nil, ms.VerifyError
	}
	return ms.GetState(key), nil
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
	*MockState
}

var _ CallContext = (*MockCallContext)(nil)

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
