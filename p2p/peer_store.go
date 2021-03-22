// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package p2p

import (
	"sync"

	"github.com/aungmawjj/juria-blockchain/core"
)

type PeerStore struct {
	peers map[string]*Peer
	mtx   sync.RWMutex
}

func NewPeerStore() *PeerStore {
	return &PeerStore{
		peers: make(map[string]*Peer),
	}
}

func (s *PeerStore) Load(pubKey *core.PublicKey) *Peer {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	return s.peers[pubKey.String()]
}

func (s *PeerStore) Store(p *Peer) *Peer {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.peers[p.PublicKey().String()] = p
	return p
}

func (s *PeerStore) Delete(pubKey *core.PublicKey) *Peer {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	p := s.peers[pubKey.String()]
	delete(s.peers, pubKey.String())
	return p
}

func (s *PeerStore) List() []*Peer {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	peers := make([]*Peer, 0, len(s.peers))
	for _, p := range s.peers {
		peers = append(peers, p)
	}
	return peers
}

func (s *PeerStore) LoadOrStore(p *Peer) (actual *Peer, loaded bool) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if actual, loaded = s.peers[p.PublicKey().String()]; loaded {
		return actual, loaded
	}
	s.peers[p.PublicKey().String()] = p
	return p, false
}
