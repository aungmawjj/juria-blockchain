// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package cluster

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"syscall"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/node"
	"github.com/multiformats/go-multiaddr"
)

type RemoteFactoryParams struct {
	JuriaPath string
	NodeCount int
	Port      int
	ApiPort   int
	Debug     bool

	LoginName string // e.g ubuntu
	KeySSH    string
	HostsPath string // file path to host ip addresses

	WorkDir       string
	RemoteWorkDir string
}

type remoteFactory struct {
	params      RemoteFactoryParams
	templateDir string
	hosts       []string
	dirs        []string
}

var _ ClusterFactory = (*remoteFactory)(nil)

func NewRemoteFactory(params RemoteFactoryParams) (ClusterFactory, error) {
	ftry := &remoteFactory{
		params: params,
	}
	if err := ftry.setup(); err != nil {
		return nil, err
	}
	return ftry, nil
}

func (ftry *remoteFactory) setup() error {
	ftry.templateDir = path.Join(ftry.params.WorkDir, "cluster_template")
	err := os.RemoveAll(ftry.templateDir) // no error if path not exist
	if err != nil {
		return err
	}
	if err := os.Mkdir(ftry.templateDir, 0755); err != nil {
		return err
	}
	ftry.hosts, err = ftry.readHosts()
	if err != nil {
		return err
	}
	if err := ftry.createRemoteDir(); err != nil {
		return err
	}
	if err := ftry.sendJuria(); err != nil {
		return err
	}
	ftry.dirs = make([]string, ftry.params.NodeCount)
	keys := make([]*core.PrivateKey, ftry.params.NodeCount)
	addrs := make([]multiaddr.Multiaddr, ftry.params.NodeCount)
	vlds := make([]node.Validator, ftry.params.NodeCount)

	// create validator infos (pubkey + addr)
	for i := 0; i < ftry.params.NodeCount; i++ {
		ftry.dirs[i] = path.Join(ftry.templateDir, strconv.Itoa(i))

		keys[i] = core.GenerateKey(nil)
		addr, err := multiaddr.NewMultiaddr(
			fmt.Sprintf("/ip4/%s/tcp/%d", ftry.hosts[i], ftry.params.Port))
		if err != nil {
			return err
		}
		addrs[i] = addr
		vlds[i] = node.Validator{
			PubKey: keys[i].PublicKey().Bytes(),
			Addr:   addrs[i].String(),
		}
	}

	// setup workdir for each node
	for i := 0; i < ftry.params.NodeCount; i++ {
		os.Mkdir(ftry.dirs[i], 0755)
		if err := writeNodeKey(ftry.dirs[i], keys[i]); err != nil {
			return err
		}
		if err := writeValidatorsFile(ftry.dirs[i], vlds); err != nil {
			return err
		}
	}
	return nil
}

func (ftry *remoteFactory) readHosts() ([]string, error) {
	raw, err := ioutil.ReadFile(ftry.params.HostsPath)
	if err != nil {
		return nil, err
	}
	hosts := strings.Split(string(raw), "\n")
	if len(hosts) < ftry.params.NodeCount {
		return nil, fmt.Errorf("not enough hosts, %d | %d", len(hosts))
	}
	return hosts, nil
}

func (ftry *remoteFactory) createRemoteDir() error {
	for i := 0; i < ftry.params.NodeCount; i++ {
		cmd := exec.Command("ssh",
			"-i", ftry.params.KeySSH,
			"-u", ftry.params.LoginName,
			ftry.hosts[i],
			fmt.Sprintf("'mkdir %s'", ftry.params.RemoteWorkDir),
		)
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

func (ftry *remoteFactory) sendJuria() error {
	for i := 0; i < ftry.params.NodeCount; i++ {
		cmd := exec.Command("scp",
			"-i", ftry.params.KeySSH,
			"-u", ftry.params.LoginName,
			ftry.params.JuriaPath,
			ftry.hosts[i]+":"+ftry.params.RemoteWorkDir,
		)
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

func (ftry *remoteFactory) sendTemplate() error {
	for i := 0; i < ftry.params.NodeCount; i++ {
		cmd := exec.Command("scp",
			"-i", ftry.params.KeySSH,
			"-u", ftry.params.LoginName,
			"-r", ftry.dirs[i],
			ftry.hosts[i]+":"+ftry.params.RemoteWorkDir+"/template",
		)
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

func (ftry *remoteFactory) SetupCluster(name string) (*Cluster, error) {
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
	// create remoteNodes
	for i := 0; i < ftry.params.NodeCount; i++ {
		nodes[i] = &remoteNode{
			juriaPath: ftry.params.JuriaPath,
			datadir:   path.Join(clusterDir, strconv.Itoa(i)),
			port:      ftry.params.Port,
			apiPort:   ftry.params.ApiPort,
			debug:     ftry.params.Debug,
		}
	}
	return &Cluster{nodes}, nil
}

type remoteNode struct {
	juriaPath string

	datadir string
	port    int
	apiPort int
	debug   bool

	running bool
	cmd     *exec.Cmd
	logFile *os.File
}

var _ Node = (*remoteNode)(nil)

func (node *remoteNode) Start() error {
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

func (node *remoteNode) Stop() {
	if !node.running {
		return
	}
	node.running = false
	syscall.Kill(node.cmd.Process.Pid, syscall.SIGTERM)
	node.logFile.Close()
}

func (node *remoteNode) GetEndpoint() string {
	return fmt.Sprintf("http://127.0.0.1:%d", node.apiPort)
}
