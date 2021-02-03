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

func TestPeer_ReadWrite(t *testing.T) {

	assert := assert.New(t)

	var p Peer
	pb := &peerBase{}

	type rwcPipe struct {
		io.Reader
		io.Writer
		io.Closer
	}
	r, w := io.Pipe()
	rwc := &rwcPipe{r, w, r}

	p = newPeerConnected(pb, rwc, func(p Peer) {})

	msg := []byte("message")

	sub, _ := p.Observe()
	go sub.Listen(func(e emitter.Event) {
		assert.EqualValues(msg, e)
	})

	p.Write(msg)
	time.Sleep(time.Millisecond)
}
