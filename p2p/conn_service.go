// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package p2p

// ConnService creates and maintains peer connections and handles read/write messages
type ConnService struct {
	store *peerStore
}

// NewConnService creates a ConnService
func NewConnService() *ConnService {
	return &ConnService{
		store: newPeerStore(),
	}
}

// AddPeer adds a new peer
func (cs *ConnService) AddPeer(pi *PeerInfo) {
	p := newPeerDisconnected(pi, func(p Peer) {})
	cs.store.LoadOrStore(p.String(), p)
}

// GetPeer returns peer with the given public key string
func (cs *ConnService) GetPeer(key string) Peer {
	return cs.store.Load(key)
}
