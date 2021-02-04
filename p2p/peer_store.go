// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package p2p

import "sync"

type peerStore struct {
	peers map[string]Peer
	mtx   sync.RWMutex
}

func newPeerStore() *peerStore {
	return &peerStore{
		peers: make(map[string]Peer),
	}
}

func (s *peerStore) Load(key string) Peer {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	return s.peers[key]
}

func (s *peerStore) Store(key string, p Peer) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.peers[key] = p
}

func (s *peerStore) Delete(key string) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	delete(s.peers, key)
}

func (s *peerStore) List() []Peer {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	peers := make([]Peer, 0, len(s.peers))
	for _, p := range s.peers {
		peers = append(peers, p)
	}
	return peers
}

func (s *peerStore) LoadOrStore(key string, p Peer) (actual Peer, loaded bool) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if actual, loaded = s.peers[key]; loaded {
		return actual, loaded
	}
	s.peers[key] = p
	return p, false
}
