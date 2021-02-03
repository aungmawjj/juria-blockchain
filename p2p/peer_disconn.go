// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package p2p

import (
	"fmt"

	"github.com/aungmawjj/juria-blockchain/emitter"
)

type peerDisconnected struct {
	PeerInfo
	onConnect func(p Peer)
}

// make sure peerDisconnected implements Peer interface
var _ Peer = (*peerDisconnected)(nil)

func newPeerDisconnected(peerInfo *PeerInfo, onConnect func(p Peer)) *peerDisconnected {
	return &peerDisconnected{
		PeerInfo:  *peerInfo,
		onConnect: onConnect,
	}
}

func (p *peerDisconnected) Status() PeerStatus {
	return PeerStatusDisconnected
}

func (p *peerDisconnected) Connect() {
	p.onConnect(p)
}

func (p *peerDisconnected) Disconnect() {
	// do nothing
}

func (p *peerDisconnected) Write(msg []byte) error {
	return fmt.Errorf("can't write! peer disconnected")
}

func (p *peerDisconnected) Observe() (*emitter.Subscription, error) {
	return nil, fmt.Errorf("can't observe! peer disconnected")
}

func (p *peerDisconnected) Block() {
	// TODO block peer
}
