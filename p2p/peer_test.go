// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package p2p

import (
	"io"
	"testing"
	"time"

	"github.com/aungmawjj/juria-blockchain/emitter"
	"github.com/stretchr/testify/assert"
)

type rwcPipe struct {
	io.Reader
	io.Writer
	io.Closer
}

func newRWCPipe() *rwcPipe {
	r, w := io.Pipe()
	return &rwcPipe{r, w, r}
}

func TestPeer_ReadWrite(t *testing.T) {
	assert := assert.New(t)
	var p Peer

	rwc := newRWCPipe()
	p = newPeerConnected(&PeerInfo{}, rwc, func(p Peer) {})

	msg := []byte("message")

	sub, _ := p.SubscribeMsg()
	go sub.Listen(func(e emitter.Event) {
		assert.EqualValues(msg, e)
	})
	assert.NoError(p.Write(msg))

	p = newPeerConnecting(&PeerInfo{})
	assert.Error(p.Write(msg))

	p = newPeerDisconnected(&PeerInfo{}, func(p Peer) {})
	assert.Error(p.Write(msg))

	time.Sleep(time.Millisecond)
}
