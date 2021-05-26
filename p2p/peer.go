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
	"github.com/aungmawjj/juria-blockchain/logger"
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

const (
	// message size limit in bytes (~100 MB)
	// to avoid out of memory allocation for reading next message
	MessageSizeLimit uint32 = 100000000
)

// Peer type
type Peer struct {
	pubKey *core.PublicKey
	addr   multiaddr.Multiaddr
	status PeerStatus

	rwc     io.ReadWriteCloser
	emitter *emitter.Emitter

	mtxRWC    sync.RWMutex
	mtxStatus sync.RWMutex
	mtxWrite  sync.Mutex
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

// Status gogoc
func (p *Peer) Status() PeerStatus {
	p.mtxStatus.RLock()
	defer p.mtxStatus.RUnlock()

	return p.status
}

// Disconnect gogoc
func (p *Peer) Disconnect() error {
	p.mtxStatus.Lock()
	defer p.mtxStatus.Unlock()

	if p.status == PeerStatusConnected {
		logger.I().Infow("peer disconnected", "addr", p.addr)
	}
	p.status = PeerStatusDisconnected
	rwc := p.getRWC()
	if rwc != nil {
		return rwc.Close()
	}
	return nil
}

// SetConnecting gogoc
func (p *Peer) SetConnecting() error {
	p.mtxStatus.Lock()
	defer p.mtxStatus.Unlock()

	if p.status != PeerStatusDisconnected {
		return fmt.Errorf("Status must be disconnected")
	}
	p.status = PeerStatusConnecting
	return nil
}

// OnConnected gogoc
func (p *Peer) OnConnected(rwc io.ReadWriteCloser) {
	p.mtxStatus.Lock()
	defer p.mtxStatus.Unlock()

	logger.I().Infow("peer connected", "addr", p.addr)
	p.status = PeerStatusConnected
	p.setRWC(rwc)
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
	if size > MessageSizeLimit {
		return nil, fmt.Errorf("big message size %d", size)
	}
	return p.readFixedSize(size)
}

func (p *Peer) readFixedSize(size uint32) ([]byte, error) {
	b := make([]byte, size)
	_, err := io.ReadFull(p.getRWC(), b)
	return b, err
}

// WriteMsg gogoc
func (p *Peer) WriteMsg(msg []byte) error {
	p.mtxWrite.Lock()
	defer p.mtxWrite.Unlock()

	if p.Status() != PeerStatusConnected {
		return fmt.Errorf("Peer not connected")
	}
	return p.write(msg)
}

func (p *Peer) write(b []byte) error {
	payload := make([]byte, 4, 4+len(b))
	binary.BigEndian.PutUint32(payload, uint32(len(b)))
	payload = append(payload, b...)

	_, err := p.getRWC().Write(payload)
	return err
}

// SubscribeMsg gogoc
func (p *Peer) SubscribeMsg() *emitter.Subscription {
	return p.emitter.Subscribe(10)
}

func (p *Peer) setRWC(rwc io.ReadWriteCloser) {
	p.mtxRWC.Lock()
	defer p.mtxRWC.Unlock()
	p.rwc = rwc
}

func (p *Peer) getRWC() io.ReadWriteCloser {
	p.mtxRWC.RLock()
	defer p.mtxRWC.RUnlock()
	return p.rwc
}
