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
	"github.com/multiformats/go-multiaddr"
)

type Validator struct {
	PubKey []byte
	Addr   string
}

const (
	NodekeyFile    = "nodekey"
	ValidatorsFile = "validators.json"
)

func readNodeKey(datadir string) (*core.PrivateKey, error) {
	b, err := ioutil.ReadFile(path.Join(datadir, NodekeyFile))
	if err != nil {
		return nil, fmt.Errorf("cannot read %s %w", NodekeyFile, err)
	}
	return core.NewPrivateKey(b)
}

func readValidators(datadir string) ([]*core.PublicKey, []multiaddr.Multiaddr, error) {
	f, err := os.Open(path.Join(datadir, ValidatorsFile))
	if err != nil {
		return nil, nil, fmt.Errorf("cannot read %s %w", ValidatorsFile, err)
	}
	defer f.Close()

	var vlds []Validator
	if err := json.NewDecoder(f).Decode(&vlds); err != nil {
		return nil, nil, fmt.Errorf("cannot parse validators json %w", err)
	}

	vldKeys := make([]*core.PublicKey, len(vlds))
	vldAddrs := make([]multiaddr.Multiaddr, len(vlds))

	for i, vld := range vlds {
		vldKeys[i], err = core.NewPublicKey(vld.PubKey)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid public key %w", err)
		}
		vldAddrs[i], err = multiaddr.NewMultiaddr(vld.Addr)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid multiaddr %w", err)
		}
	}
	return vldKeys, vldAddrs, nil
}
