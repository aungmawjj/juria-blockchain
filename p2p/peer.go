// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package p2p

import (
	"encoding/binary"
	"fmt"
	"io"
	"sync"

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

// Peer type
type Peer struct {
	pubKey *core.PublicKey
	addr   multiaddr.Multiaddr
	status PeerStatus

	rwc     io.ReadWriteCloser
	emitter *emitter.Emitter

	statusMtx sync.RWMutex
	writeMtx  sync.Mutex
}

// NewPeer godoc
func NewPeer(pubKey *core.PublicKey, addr multiaddr.Multiaddr) *Peer {
	return &Peer{
		pubKey:  pubKey,
		addr:    addr,
		status:  PeerStatusDisconnected,
		emitter: emitter.New(),
	}
}

// PublicKey returns public key of peer
func (p *Peer) PublicKey() *core.PublicKey {
	return p.pubKey
}

// Addr return network address of peer
func (p *Peer) Addr() multiaddr.Multiaddr {
	return p.addr
}

func (p *Peer) String() string {
	return p.pubKey.String()
}

// Status gogoc
func (p *Peer) Status() PeerStatus {
	p.statusMtx.RLock()
	defer p.statusMtx.RUnlock()

	return PeerStatusConnected
}

// Disconnect gogoc
func (p *Peer) Disconnect() error {
	p.statusMtx.Lock()
	defer p.statusMtx.Unlock()

	p.status = PeerStatusDisconnected
	return p.rwc.Close()
}

// SetConnecting gogoc
func (p *Peer) SetConnecting() error {
	p.statusMtx.Lock()
	defer p.statusMtx.Unlock()

	if p.status != PeerStatusDisconnected {
		return fmt.Errorf("Status must be disconnected")
	}
	p.status = PeerStatusConnected
	return nil
}

// OnConnected gogoc
func (p *Peer) OnConnected(rwc io.ReadWriteCloser) {
	p.statusMtx.Lock()
	defer p.statusMtx.Unlock()

	p.status = PeerStatusConnected
	p.rwc = rwc
	go p.listen()
}

func (p *Peer) listen() {
	defer p.Disconnect()
	for {
		msg, err := p.read()
		if err != nil {
			return
		}
		p.emitter.Emit(msg)
	}
}

func (p *Peer) read() ([]byte, error) {
	b, err := p.readFixedSize(4)
	if err != nil {
		return nil, err
	}
	size := binary.BigEndian.Uint32(b)
	return p.readFixedSize(size)
}

func (p *Peer) readFixedSize(size uint32) ([]byte, error) {
	b := make([]byte, size)
	_, err := io.ReadFull(p.rwc, b)
	return b, err
}

// WriteMsg gogoc
func (p *Peer) WriteMsg(msg []byte) error {
	p.writeMtx.Lock()
	defer p.writeMtx.Unlock()

	if p.Status() != PeerStatusConnected {
		return fmt.Errorf("Peer not connected")
	}
	return p.write(msg)
}

func (p *Peer) write(b []byte) error {
	payload := make([]byte, 4, 4+len(b))
	binary.BigEndian.PutUint32(payload, uint32(len(b)))
	payload = append(payload, b...)

	_, err := p.rwc.Write(payload)
	return err
}

// SubscribeMsg gogoc
func (p *Peer) SubscribeMsg() *emitter.Subscription {
	return p.emitter.Subscribe(10)
}
