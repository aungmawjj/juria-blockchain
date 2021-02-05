// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package p2p

import (
	"encoding/binary"
	"io"
	"sync"

	"github.com/aungmawjj/juria-blockchain/emitter"
)

type peerConnected struct {
	PeerInfo
	rwc          io.ReadWriteCloser
	emitter      *emitter.Emitter
	mtx          sync.Mutex
	onDisconnect func(p Peer)
}

var _ Peer = (*peerConnected)(nil)

func newPeerConnected(peerInfo *PeerInfo, rwc io.ReadWriteCloser, onDisconnect func(p Peer)) *peerConnected {
	p := &peerConnected{
		PeerInfo:     *peerInfo,
		rwc:          rwc,
		emitter:      emitter.New(),
		onDisconnect: onDisconnect,
	}
	go p.listen()
	return p
}

func (p *peerConnected) listen() {
	defer p.Disconnect()
	for {
		msg, err := p.read()
		if err != nil {
			return
		}
		p.emitter.Emit(msg)
	}
}

func (p *peerConnected) read() ([]byte, error) {
	b, err := p.readFixedSize(4)
	if err != nil {
		return nil, err
	}
	size := binary.BigEndian.Uint32(b)
	return p.readFixedSize(size)
}

func (p *peerConnected) readFixedSize(size uint32) ([]byte, error) {
	b := make([]byte, size)
	_, err := io.ReadFull(p.rwc, b)
	return b, err
}

func (p *peerConnected) Status() PeerStatus {
	return PeerStatusConnected
}

func (p *peerConnected) Connect() {
	// do nothing
}

func (p *peerConnected) Disconnect() {
	p.rwc.Close()
	p.onDisconnect(p)
}

func (p *peerConnected) Write(msg []byte) error {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	payload := make([]byte, 4, 4+len(msg))
	binary.BigEndian.PutUint32(payload, uint32(len(msg)))
	payload = append(payload, msg...)

	_, err := p.rwc.Write(payload)
	return err
}

func (p *peerConnected) SubscribeMsg() (*emitter.Subscription, error) {
	return p.emitter.Subscribe(10), nil
}

func (p *peerConnected) Block() {
	// TODO block peer
}
