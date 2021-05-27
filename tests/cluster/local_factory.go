// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package cluster

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/node"
	"github.com/multiformats/go-multiaddr"
)

type LocalFactoryParams struct {
	JuriaPath string
	WorkDir   string
	NodeCount int
	PortN0    int // node zero port
	ApiPortN0 int // node zero api port
}

type localFactory struct {
	params      LocalFactoryParams
	templateDir string
}

var _ ClusterFactory = (*localFactory)(nil)

func NewLocalFactory(params LocalFactoryParams) (ClusterFactory, error) {
	ftry := &localFactory{
		params: params,
	}
	if err := ftry.setup(); err != nil {
		return nil, err
	}
	return ftry, nil
}

func (ftry *localFactory) setup() error {
	ftry.templateDir = path.Join(ftry.params.WorkDir, "cluster_template")
	err := os.RemoveAll(ftry.templateDir) // no error if path not exist
	if err != nil {
		return err
	}
	if err := os.Mkdir(ftry.templateDir, 0755); err != nil {
		return err
	}

	keys := make([]*core.PrivateKey, ftry.params.NodeCount)
	ports := make([]string, ftry.params.NodeCount)
	apiPorts := make([]string, ftry.params.NodeCount)
	addrs := make([]multiaddr.Multiaddr, ftry.params.NodeCount)
	vlds := make([]node.Validator, ftry.params.NodeCount)
	dirs := make([]string, ftry.params.NodeCount)

	// create validator infos (pubkey + addr)
	for i := 0; i < ftry.params.NodeCount; i++ {
		keys[i] = core.GenerateKey(nil)
		ports[i] = strconv.Itoa(ftry.params.PortN0 + i)
		apiPorts[i] = strconv.Itoa(ftry.params.ApiPortN0 + i)
		addr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%s", ports[i]))
		if err != nil {
			return err
		}
		addrs[i] = addr
		vlds[i] = node.Validator{
			PubKey: keys[i].PublicKey().Bytes(),
			Addr:   addrs[i].String(),
		}
		dirs[i] = path.Join(ftry.templateDir, strconv.Itoa(i))
	}

	// setup workdir for each node
	for i := 0; i < ftry.params.NodeCount; i++ {
		os.Mkdir(dirs[i], 0755)
		if err := writeNodeKey(dirs[i], keys[i]); err != nil {
			return err
		}
		if err := writeValidatorsFile(dirs[i], vlds); err != nil {
			return err
		}
	}
	return nil
}

func (ftry *localFactory) GetCluster(name string) (*Cluster, error) {
	clusterDir := path.Join(ftry.params.WorkDir, name)
	err := os.RemoveAll(clusterDir) // no error if path not exist
	if err != nil {
		return nil, err
	}
	err = exec.Command("cp", "-r", ftry.templateDir, clusterDir).Run()
	if err != nil {
		return nil, err
	}

	nodes := make([]Node, ftry.params.NodeCount)
	// create localNodes
	for i := 0; i < ftry.params.NodeCount; i++ {
		nodes[i] = &localNode{
			juriaPath: ftry.params.JuriaPath,
			datadir:   path.Join(clusterDir, strconv.Itoa(i)),
			port:      ftry.params.PortN0 + i,
			apiPort:   ftry.params.ApiPortN0 + i,
		}
	}
	return &Cluster{nodes}, nil
}
