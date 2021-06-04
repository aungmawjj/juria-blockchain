// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package node

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/p2p"
	"github.com/multiformats/go-multiaddr"
)

type Peer struct {
	PubKey []byte
	Addr   string
}

type Genesis struct {
	Validators [][]byte
}

const (
	NodekeyFile = "nodekey"
	GenesisFile = "genesis.json"
	PeersFile   = "peers.json"
)

func readNodeKey(datadir string) (*core.PrivateKey, error) {
	b, err := ioutil.ReadFile(path.Join(datadir, NodekeyFile))
	if err != nil {
		return nil, fmt.Errorf("cannot read %s, %w", NodekeyFile, err)
	}
	return core.NewPrivateKey(b)
}

func readGenesis(datadir string) (*Genesis, error) {
	f, err := os.Open(path.Join(datadir, GenesisFile))
	if err != nil {
		return nil, fmt.Errorf("cannot read %s, %w", GenesisFile, err)
	}
	defer f.Close()

	genesis := new(Genesis)
	if err := json.NewDecoder(f).Decode(&genesis); err != nil {
		return nil, fmt.Errorf("cannot parse %s, %w", GenesisFile, err)

	}
	return genesis, nil
}

func readPeers(datadir string) ([]*p2p.Peer, error) {
	f, err := os.Open(path.Join(datadir, PeersFile))
	if err != nil {
		return nil, fmt.Errorf("cannot read %s, %w", PeersFile, err)
	}
	defer f.Close()

	var raws []Peer
	if err := json.NewDecoder(f).Decode(&raws); err != nil {
		return nil, fmt.Errorf("cannot parse %s, %w", PeersFile, err)
	}

	peers := make([]*p2p.Peer, len(raws))

	for i, r := range raws {
		pubKey, err := core.NewPublicKey(r.PubKey)
		if err != nil {
			return nil, fmt.Errorf("invalid public key %w", err)
		}
		addr, err := multiaddr.NewMultiaddr(r.Addr)
		if err != nil {
			return nil, fmt.Errorf("invalid multiaddr %w", err)
		}
		peers[i] = p2p.NewPeer(pubKey, addr)
	}
	return peers, nil
}
