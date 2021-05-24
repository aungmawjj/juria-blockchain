// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package p2p

import (
	"context"
	"errors"
	"time"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/multiformats/go-multiaddr"
)

const protocolID = "/single_pid"

type Host struct {
	privKey   *core.PrivateKey
	localAddr multiaddr.Multiaddr

	peerStore *PeerStore
	libHost   host.Host

	onAddedPeer func(peer *Peer)

	reconnectInterval time.Duration
}

func NewHost(privKey *core.PrivateKey, localAddr multiaddr.Multiaddr) (*Host, error) {
	host := new(Host)
	host.privKey = privKey
	host.localAddr = localAddr
	host.peerStore = NewPeerStore()

	libHost, err := host.newLibHost()
	if err != nil {
		return nil, err
	}
	host.libHost = libHost
	host.libHost.SetStreamHandler(protocolID, host.handleStream)
	host.reconnectInterval = 5 * time.Second
	go host.reconnectLoop()
	return host, nil
}

func (host *Host) newLibHost() (host.Host, error) {
	priv, err := crypto.UnmarshalEd25519PrivateKey(host.privKey.Bytes())
	if err != nil {
		return nil, err
	}
	return libp2p.New(
		context.Background(),
		libp2p.Identity(priv),
		libp2p.ListenAddrs(host.localAddr),
	)
}

func (host *Host) handleStream(s network.Stream) {
	pubKey, err := getRemotePublicKey(s)
	if err != nil {
		return
	}
	peer, loaded := host.peerStore.LoadOrStore(NewPeer(pubKey, s.Conn().RemoteMultiaddr()))
	if !loaded && host.onAddedPeer != nil {
		go host.onAddedPeer(peer)
	}
	if err := peer.SetConnecting(); err != nil {
		s.Close()
		return
	}
	peer.OnConnected(s)
}

func (host *Host) reconnectLoop() {
	for range time.Tick(host.reconnectInterval) {
		peers := host.peerStore.List()
		for _, peer := range peers {
			go host.connectPeer(peer)
		}
	}
}

func (host *Host) connectPeer(peer *Peer) {
	if err := peer.SetConnecting(); err != nil { // prevent simultaneous connections from both hosts
		return
	}
	s, err := host.newStream(peer)
	if err != nil {
		peer.Disconnect()
		return
	}
	peer.OnConnected(s)
}

func (host *Host) newStream(peer *Peer) (network.Stream, error) {
	id, err := getIDFromPublicKey(peer.PublicKey())
	if err != nil {
		return nil, err
	}
	host.libHost.Peerstore().AddAddr(id, peer.Addr(), peerstore.PermanentAddrTTL)
	return host.libHost.NewStream(context.Background(), id, protocolID)
}

func (host *Host) AddPeer(peer *Peer) {
	peer, loaded := host.peerStore.LoadOrStore(peer)
	if !loaded && host.onAddedPeer != nil {
		go host.onAddedPeer(peer)
	}
	go host.connectPeer(peer)
}

func (host *Host) SetPeerAddedHandler(fn func(peer *Peer)) {
	host.onAddedPeer = fn
}

func (host *Host) PeerStore() *PeerStore {
	return host.peerStore
}

func getRemotePublicKey(s network.Stream) (*core.PublicKey, error) {
	if _, ok := s.Conn().RemotePublicKey().(*crypto.Ed25519PublicKey); !ok {
		return nil, errors.New("invalid pubKey type")
	}
	b, err := s.Conn().RemotePublicKey().Raw()
	if err != nil {
		return nil, err
	}
	return core.NewPublicKey(b)
}

func getIDFromPublicKey(pubKey *core.PublicKey) (peer.ID, error) {
	var id peer.ID
	if pubKey == nil {
		return id, errors.New("nil peer pubkey")
	}
	key, err := crypto.UnmarshalEd25519PublicKey(pubKey.Bytes())
	if err != nil {
		return id, err
	}
	return peer.IDFromPublicKey(key)
}
