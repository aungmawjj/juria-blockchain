// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package cluster

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"syscall"

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
	Debug     bool
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
	addrs := make([]multiaddr.Multiaddr, ftry.params.NodeCount)
	vlds := make([]node.Validator, ftry.params.NodeCount)
	dirs := make([]string, ftry.params.NodeCount)

	// create validator infos (pubkey + addr)
	for i := 0; i < ftry.params.NodeCount; i++ {
		keys[i] = core.GenerateKey(nil)
		addr, err := multiaddr.NewMultiaddr(
			fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", ftry.params.PortN0+i))
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

func (ftry *localFactory) SetupCluster(name string) (*Cluster, error) {
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
			debug:     ftry.params.Debug,
		}
	}
	return &Cluster{nodes}, nil
}

type localNode struct {
	juriaPath string

	datadir string
	port    int
	apiPort int
	debug   bool

	running bool
	cmd     *exec.Cmd
	logFile *os.File
}

var _ Node = (*localNode)(nil)

func (node *localNode) Start() error {
	if node.running {
		return nil
	}
	f, err := os.OpenFile(path.Join(node.datadir, "log.txt"),
		os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	node.logFile = f
	node.cmd = exec.Command(node.juriaPath,
		"-d", node.datadir,
		"-p", strconv.Itoa(node.port),
		"-P", strconv.Itoa(node.apiPort),
	)
	if node.debug {
		node.cmd.Args = append(node.cmd.Args, "--debug")
	}
	node.cmd.Stderr = node.logFile
	node.cmd.Stdout = node.logFile
	node.running = true
	return node.cmd.Start()
}

func (node *localNode) Stop() {
	if !node.running {
		return
	}
	node.running = false
	syscall.Kill(node.cmd.Process.Pid, syscall.SIGTERM)
	node.logFile.Close()
}

func (node *localNode) GetEndpoint() string {
	return fmt.Sprintf("http://127.0.0.1:%d", node.apiPort)
}
