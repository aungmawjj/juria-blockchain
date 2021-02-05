// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package p2p

import (
	"fmt"

	"github.com/aungmawjj/juria-blockchain/emitter"
)

type peerConnecting struct {
	PeerInfo
}

var _ Peer = (*peerConnecting)(nil)

func newPeerConnecting(peerInfo *PeerInfo) *peerConnecting {
	return &peerConnecting{
		PeerInfo: *peerInfo,
	}
}

func (p *peerConnecting) Status() PeerStatus {
	return PeerStatusConnecting
}

func (p *peerConnecting) Connect() {
	// do nothing
}

func (p *peerConnecting) Disconnect() {
	// do nothing
}

func (p *peerConnecting) Write(msg []byte) error {
	return fmt.Errorf("can't write! peer connecting")
}

func (p *peerConnecting) SubscribeMsg() (*emitter.Subscription, error) {
	return nil, fmt.Errorf("can't observe! peer connecting")
}

func (p *peerConnecting) Block() {
	// TODO block peer
}
