// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package p2p

import (
	"io"
	"testing"
	"time"

	"github.com/aungmawjj/juria-blockchain/emitter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type rwcLoopBack struct {
	io.Reader
	io.Writer
	io.Closer
}

func newRWCLoopBack() *rwcLoopBack {
	r, w := io.Pipe()
	return &rwcLoopBack{r, w, r}
}

type MockListener struct {
	mock.Mock
}

func (m *MockListener) cb(e emitter.Event) {
	m.Called(e)
}

func TestRWCLoopBack(t *testing.T) {
	assert := assert.New(t)

	rwc := newRWCLoopBack()
	recv := make([]byte, 5)
	go func() {
		for {
			rwc.Read(recv)
		}
	}()

	sent := []byte("hello")
	rwc.Write(sent)

	time.Sleep(time.Millisecond)
	assert.EqualValues(sent, recv)

	rwc.Close()
	_, err := rwc.Write(sent)
	assert.Error(err)
}

func TestPeer_ReadWrite(t *testing.T) {
	assert := assert.New(t)
	p := NewPeer(nil, nil)

	rwc := newRWCLoopBack()
	p.OnConnected(rwc)
	sub := p.SubscribeMsg()

	msg := []byte("hello")

	mln := new(MockListener)
	mln.On("cb", mock.Anything).Once()

	go sub.Listen(mln.cb)
	assert.NoError(p.WriteMsg(msg))

	time.Sleep(time.Millisecond)

	mln.AssertExpectations(t)
}