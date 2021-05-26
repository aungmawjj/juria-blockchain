// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package p2p

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/aungmawjj/juria-blockchain/emitter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type rwcLoopBack struct {
	buf      *bytes.Buffer
	closedCh chan struct{}
	inCh     chan struct{}
}

func newRWCLoopBack() *rwcLoopBack {
	return &rwcLoopBack{
		buf:      bytes.NewBuffer(nil),
		closedCh: make(chan struct{}),
		inCh:     make(chan struct{}, 2),
	}
}

func (rwc *rwcLoopBack) Read(b []byte) (n int, err error) {
	select {
	case <-rwc.closedCh:
		return 0, io.EOF
	default:
		n, err = rwc.readBuf(b)
		if err == io.EOF {
			select {
			case <-rwc.closedCh:
			case <-rwc.inCh:
				return rwc.Read(b)
			}
		}
		return n, err
	}
}

func (rwc *rwcLoopBack) readBuf(b []byte) (int, error) {
	return rwc.buf.Read(b)
}

func (rwc *rwcLoopBack) Write(b []byte) (n int, err error) {
	select {
	case <-rwc.closedCh:
		return 0, io.EOF
	default:
	}
	n, err = rwc.buf.Write(b)
	select {
	case rwc.inCh <- struct{}{}:
	default:
	}
	return n, err
}

func (rwc *rwcLoopBack) Close() error {
	select {
	case <-rwc.closedCh:
		return io.EOF
	default:
		close(rwc.closedCh)
		rwc.buf.Reset()
		return nil
	}
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

type MockListener struct {
	mock.Mock
}

func (m *MockListener) CB(e emitter.Event) {
	m.Called(e)
}

func TestPeer_ReadWrite(t *testing.T) {
	assert := assert.New(t)
	p := NewPeer(nil, nil)

	rwc := newRWCLoopBack()
	p.OnConnected(rwc)
	sub := p.SubscribeMsg()

	msg := []byte("hello")

	mln := new(MockListener)
	mln.On("CB", msg).Once()

	go func() {
		for event := range sub.Events() {
			mln.CB(event)
		}
	}()

	assert.NoError(p.WriteMsg(msg))

	time.Sleep(time.Millisecond)

	mln.AssertExpectations(t)
}

func TestPeer_ConnStatus(t *testing.T) {
	assert := assert.New(t)
	p := NewPeer(nil, nil)

	assert.Equal(PeerStatusDisconnected, p.Status())

	rwc := newRWCLoopBack()
	p.OnConnected(rwc)

	assert.Equal(PeerStatusConnected, p.Status())

	rwc.Close()
	time.Sleep(time.Millisecond)

	assert.Equal(PeerStatusDisconnected, p.Status())

	p = NewPeer(nil, nil)
	err := p.SetConnecting()

	assert.NoError(err)
	assert.Equal(PeerStatusConnecting, p.Status())

	p.Disconnect()
	assert.Equal(PeerStatusDisconnected, p.Status())

	p.OnConnected(newRWCLoopBack())
	err = p.SetConnecting()

	assert.Error(err)
	assert.Equal(PeerStatusConnected, p.Status())
}
