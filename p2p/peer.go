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
	Info() *PeerInfo
	PublicKey() *core.PublicKey
	Addr() multiaddr.Multiaddr
	String() string
	Status() PeerStatus
	Connect()
	Disconnect()
	Write(msg []byte) error
	SubscribeMsg() (*emitter.Subscription, error)
	Block()
}

// PeerInfo type
type PeerInfo struct {
	pubKey *core.PublicKey
	addr   multiaddr.Multiaddr
}

// Info returns PeerInfo pointer
func (p PeerInfo) Info() *PeerInfo {
	return &p
}

// PublicKey of peer
func (p PeerInfo) PublicKey() *core.PublicKey {
	return p.pubKey
}

// Addr of peer
func (p PeerInfo) Addr() multiaddr.Multiaddr {
	return p.addr
}

func (p PeerInfo) String() string {
	return p.pubKey.String()
}
