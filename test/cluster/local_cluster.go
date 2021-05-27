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

type LocalClusterFactory struct {
	JuriaPath string
	Workdir   string
	NodeCount int
	PortN0    int // node zero port
	ApiPortN0 int // node zero api port
}

var _ ClusterFactory = (*LocalClusterFactory)(nil)

func (ftry *LocalClusterFactory) GetCluster(name string) Cluster {
	return &localCluster{
		juriaPath: ftry.JuriaPath,
		workdir:   path.Join(ftry.Workdir, name),
		count:     ftry.NodeCount,
		portN0:    ftry.PortN0,
		apiPortN0: ftry.ApiPortN0,
	}
}

type localCluster struct {
	juriaPath string
	workdir   string
	count     int
	portN0    int // node zero port
	apiPortN0 int // node zero api port

	nodes []*localNode
}

var _ Cluster = (*localCluster)(nil)

func (lcc *localCluster) Setup() error {
	err := os.RemoveAll(lcc.workdir) // no error if path not exist
	if err != nil {
		return err
	}
	if err := os.Mkdir(lcc.workdir, 0755); err != nil {
		return err
	}

	keys := make([]*core.PrivateKey, lcc.count)
	ports := make([]string, lcc.count)
	apiPorts := make([]string, lcc.count)
	addrs := make([]multiaddr.Multiaddr, lcc.count)
	vlds := make([]node.Validator, lcc.count)
	dirs := make([]string, lcc.count)

	// create validator infos (pubkey + addr)
	for i := 0; i < lcc.count; i++ {
		keys[i] = core.GenerateKey(nil)
		ports[i] = strconv.Itoa(lcc.portN0 + i)
		apiPorts[i] = strconv.Itoa(lcc.apiPortN0 + i)
		addr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%s", ports[i]))
		if err != nil {
			return err
		}
		addrs[i] = addr
		vlds[i] = node.Validator{
			PubKey: keys[i].PublicKey().Bytes(),
			Addr:   addrs[i].String(),
		}
		dirs[i] = path.Join(lcc.workdir, fmt.Sprintf("%0d", i))
	}

	// setup workdir for each node
	for i := 0; i < lcc.count; i++ {
		os.Mkdir(dirs[i], 0755)
		if err := lcc.writeNodeKey(dirs[i], keys[i]); err != nil {
			return err
		}
		if err := lcc.writeValidatorsFile(dirs[i], vlds); err != nil {
			return err
		}
	}

	lcc.nodes = make([]*localNode, lcc.count)
	// create localNodes
	for i := 0; i < lcc.count; i++ {
		lcc.nodes[i] = &localNode{
			juriaPath: lcc.juriaPath,
			datadir:   dirs[i],
			port:      ports[i],
		}
	}

	return nil
}

func (lcc *localCluster) writeNodeKey(nodedir string, key *core.PrivateKey) error {
	f, err := os.Create(path.Join(nodedir, node.NodekeyFile))
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(key.Bytes())
	return err
}

func (lcc *localCluster) writeValidatorsFile(nodedir string, vlds []node.Validator) error {
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

func (lcc *localCluster) Stop() error {
	for _, node := range lcc.nodes {
		node.Stop() // ignore node stop error
	}
	return nil
}

func (lcc *localCluster) GetNode(idx int) Node {
	if idx >= len(lcc.nodes) || idx < 0 {
		return nil
	}
	return lcc.nodes[idx]
}
