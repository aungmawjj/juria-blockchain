// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package p2p

import (
	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/emitter"
	"github.com/multiformats/go-multiaddr"
)

// PeerStatus type
type PeerStatus int8

// PeerStatus
const (
	PeerStatusDisconnected PeerStatus = iota
	PeerStatusConnecting
	PeerStatusConnected
	PeerStatusBlocked
)

// Peer repreents a remote peer connection
type Peer interface {
	PublicKey() *core.PublicKey
	Addr() multiaddr.Multiaddr
	Status() PeerStatus
	Connect()
	Disconnect()
	Write(msg []byte) error
	Observe() (*emitter.Subscription, error)
	Block()
}

type peerBase struct {
	pubKey *core.PublicKey
	addr   multiaddr.Multiaddr
}

func (p peerBase) PublicKey() *core.PublicKey {
	return p.pubKey
}

func (p peerBase) Addr() multiaddr.Multiaddr {
	return p.addr
}
