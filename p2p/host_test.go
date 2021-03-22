// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package p2p

import (
	"testing"
	"time"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
)

func TestHost(t *testing.T) {
	assert := assert.New(t)

	priv1 := core.GenerateKey(nil)
	priv2 := core.GenerateKey(nil)

	addr1, _ := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/25001")
	addr2, _ := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/25002")

	host1, err := NewHost(priv1, addr1)
	if !assert.NoError(err) {
		return
	}
	host2, err := NewHost(priv2, addr2)
	if !assert.NoError(err) {
		return
	}

	addedPeerCalls := 0
	onAddedPeer := func(peer *Peer) {
		addedPeerCalls++
	}
	host1.SetPeerAddedHandler(onAddedPeer)
	host2.SetPeerAddedHandler(onAddedPeer)

	host1.AddPeer(NewPeer(priv2.PublicKey(), addr2))

	time.Sleep(5 * time.Millisecond)

	assert.Equal(2, addedPeerCalls)

	p1 := host2.PeerStore().Load(priv1.PublicKey())
	p2 := host1.PeerStore().Load(priv2.PublicKey())

	if !assert.NotNil(p1) {
		return
	}
	if !assert.NotNil(p2) {
		return
	}
	assert.Equal(PeerStatusConnected, p1.Status())
	assert.Equal(PeerStatusConnected, p2.Status())

	// wait message from host2
	s1 := p1.SubscribeMsg()
	var recv1 []byte
	go func() {
		for e := range s1.Events() {
			recv1 = e.([]byte)
		}
	}()

	// send message from host1
	msg := []byte("hello")
	p2.WriteMsg(msg)

	time.Sleep(5 * time.Millisecond)

	assert.Equal(msg, recv1)

	// wait message from host1
	s2 := p2.SubscribeMsg()
	var recv2 []byte
	go func() {
		for e := range s2.Events() {
			recv2 = e.([]byte)
		}
	}()

	// send message from host2
	p1.WriteMsg(msg)

	time.Sleep(5 * time.Millisecond)

	assert.Equal(msg, recv2)

	priv3 := core.GenerateKey(nil)
	host1.AddPeer(NewPeer(priv3.PublicKey(), addr2)) // invalid key

	time.Sleep(5 * time.Millisecond)

	p3 := host1.PeerStore().Load(priv3.PublicKey())
	if assert.NotNil(p3) {
		assert.Equal(PeerStatusDisconnected, p3.Status())
	}

	addr3, _ := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/25003")
	host2.AddPeer(NewPeer(priv3.PublicKey(), addr3)) // not reachable host

	time.Sleep(5 * time.Millisecond)

	p4 := host2.PeerStore().Load(priv3.PublicKey())
	if assert.NotNil(p4) {
		assert.Equal(PeerStatusDisconnected, p4.Status())
	}
}
