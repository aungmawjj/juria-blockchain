// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package cluster

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/node"
	"github.com/multiformats/go-multiaddr"
)

type LocalClusterParams struct {
	JuriaPath string
	Workdir   string
	NodeCount int
	PortN0    int // node zero port
	ApiPortN0 int // node zero api port
}

type localCluster struct {
	nodes []*localNode
}

var _ Cluster = (*localCluster)(nil)

func SetupLocalCluster(params LocalClusterParams) (Cluster, error) {
	err := os.RemoveAll(params.Workdir) // no error if path not exist
	if err != nil {
		return nil, err
	}
	if err := os.Mkdir(params.Workdir, 0755); err != nil {
		return nil, err
	}

	keys := make([]*core.PrivateKey, params.NodeCount)
	ports := make([]string, params.NodeCount)
	apiPorts := make([]string, params.NodeCount)
	addrs := make([]multiaddr.Multiaddr, params.NodeCount)
	vlds := make([]node.Validator, params.NodeCount)
	dirs := make([]string, params.NodeCount)

	// create validator infos (pubkey + addr)
	for i := 0; i < params.NodeCount; i++ {
		keys[i] = core.GenerateKey(nil)
		ports[i] = strconv.Itoa(params.PortN0 + i)
		apiPorts[i] = strconv.Itoa(params.ApiPortN0 + i)
		addr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%s", ports[i]))
		if err != nil {
			return nil, err
		}
		addrs[i] = addr
		vlds[i] = node.Validator{
			PubKey: keys[i].PublicKey().Bytes(),
			Addr:   addrs[i].String(),
		}
		dirs[i] = path.Join(params.Workdir, fmt.Sprintf("%0d", i))
	}

	// setup workdir for each node
	for i := 0; i < params.NodeCount; i++ {
		os.Mkdir(dirs[i], 0755)
		if err := writeNodeKey(dirs[i], keys[i]); err != nil {
			return nil, err
		}
		if err := writeValidatorsFile(dirs[i], vlds); err != nil {
			return nil, err
		}
	}

	nodes := make([]*localNode, params.NodeCount)
	// create localNodes
	for i := 0; i < params.NodeCount; i++ {
		nodes[i] = &localNode{
			juriaPath: params.JuriaPath,
			datadir:   dirs[i],
			port:      ports[i],
		}
	}

	return &localCluster{nodes}, nil
}

func writeNodeKey(nodedir string, key *core.PrivateKey) error {
	f, err := os.Create(path.Join(nodedir, node.NodekeyFile))
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(key.Bytes())
	return err
}

func writeValidatorsFile(nodedir string, vlds []node.Validator) error {
	f, err := os.Create(path.Join(nodedir, node.ValidatorsFile))
	if err != nil {
		return err
	}
	defer f.Close()
	e := json.NewEncoder(f)
	e.SetIndent("", "  ")
	return e.Encode(vlds)
}

func (lcc *localCluster) Start() error {
	for _, node := range lcc.nodes {
		if err := node.Start(); err != nil {
			return err
		}
	}
	return nil
}

func (lcc *localCluster) Stop() {
	for _, node := range lcc.nodes {
		node.Stop()
	}
}

func (lcc *localCluster) NodeCount() int {
	return len(lcc.nodes)
}

func (lcc *localCluster) GetNode(idx int) Node {
	if idx >= len(lcc.nodes) || idx < 0 {
		return nil
	}
	return lcc.nodes[idx]
}
