// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package p2p

import "github.com/aungmawjj/juria-blockchain/emitter"

type peerConnected struct {
	peerBase
	onDisconnect func(p Peer)
}

var _ Peer = (*peerConnected)(nil)

func newPeerConnected(peerBase *peerBase, onDisconnect func(p Peer)) *peerConnected {
	return &peerConnected{
		peerBase:     *peerBase,
		onDisconnect: onDisconnect,
	}
}

func (p *peerConnected) Status() PeerStatus {
	return PeerStatusConnected
}

func (p *peerConnected) Connect() {
	// do nothing
}

func (p *peerConnected) Disconnect() {
	p.onDisconnect(p)
}

func (p *peerConnected) Write(msg []byte) error {
	// TODO write to stream
	return nil
}

func (p *peerConnected) Observe() (*emitter.Subscription, error) {
	return nil, nil
}

func (p *peerConnected) Block() {
	// TODO block peer
}
