// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package cluster

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/aungmawjj/juria-blockchain/node"
	"github.com/multiformats/go-multiaddr"
)

type RemoteFactoryParams struct {
	JuriaPath string
	WorkDir   string
	NodeCount int

	NodeConfig node.Config

	LoginName string // e.g ubuntu
	KeySSH    string
	HostsPath string // file path to host ip addresses

	RemoteWorkDir string
	SetupRequired bool
}

type RemoteFactory struct {
	params      RemoteFactoryParams
	templateDir string
	hosts       []string
}

var _ ClusterFactory = (*RemoteFactory)(nil)

func NewRemoteFactory(params RemoteFactoryParams) (*RemoteFactory, error) {
	ftry := &RemoteFactory{
		params: params,
	}
	ftry.templateDir = path.Join(ftry.params.WorkDir, "cluster_template")
	hosts, err := ftry.getHosts()
	if err != nil {
		return nil, err
	}
	ftry.hosts = hosts
	if ftry.params.SetupRequired {
		if err := ftry.setup(); err != nil {
			return nil, err
		}
	}
	return ftry, nil
}

func (ftry *RemoteFactory) setup() error {
	if err := ftry.setupRemoteDir(); err != nil {
		return err
	}
	if err := ftry.sendJuria(); err != nil {
		return err
	}
	addrs, err := ftry.makeAddrs()
	if err != nil {
		return err
	}
	keys := makeRandomKeys(ftry.params.NodeCount)
	vlds := makeValidators(keys, addrs)
	if err := setupTemplateDir(ftry.templateDir, keys, vlds); err != nil {
		return err
	}
	return ftry.sendTemplate()
}

func (ftry *RemoteFactory) makeAddrs() ([]multiaddr.Multiaddr, error) {
	addrs := make([]multiaddr.Multiaddr, ftry.params.NodeCount)
	// create validator infos (pubkey + addr)
	for i := range addrs {
		addr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d",
			ftry.hosts[i], ftry.params.NodeConfig.Port))
		if err != nil {
			return nil, err
		}
		addrs[i] = addr
	}
	return addrs, nil
}

func (ftry *RemoteFactory) getHosts() ([]string, error) {
	raw, err := ioutil.ReadFile(ftry.params.HostsPath)
	if err != nil {
		return nil, err
	}
	hosts := strings.Split(string(raw), "\n")
	if len(hosts) < ftry.params.NodeCount {
		return nil, fmt.Errorf("not enough hosts, %d | %d",
			len(hosts), ftry.params.NodeCount)
	}
	return hosts, nil
}

func (ftry *RemoteFactory) setupRemoteDir() error {
	for i := 0; i < ftry.params.NodeCount; i++ {
		ftry.setupRemoteDirOne(i)
	}
	return nil
}

func (ftry *RemoteFactory) setupRemoteDirOne(i int) error {
	cmd := exec.Command("ssh",
		"-i", ftry.params.KeySSH,
		fmt.Sprintf("%s@%s", ftry.params.LoginName, ftry.hosts[i]),
		"mkdir", ftry.params.RemoteWorkDir, ";",
		"cd", ftry.params.RemoteWorkDir, ";",
		"rm", "-r", "template", ";",
		"sudo", "killall", "juria",
	)
	return runCommand(cmd)
}

func (ftry *RemoteFactory) sendJuria() error {
	for i := 0; i < ftry.params.NodeCount; i++ {
		err := ftry.sendJuriaOne(i)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ftry *RemoteFactory) sendJuriaOne(i int) error {
	cmd := exec.Command("scp",
		"-i", ftry.params.KeySSH,
		ftry.params.JuriaPath,
		fmt.Sprintf("%s@%s:%s", ftry.params.LoginName, ftry.hosts[i],
			ftry.params.RemoteWorkDir),
	)
	return runCommand(cmd)
}

func (ftry *RemoteFactory) sendTemplate() error {
	for i := 0; i < ftry.params.NodeCount; i++ {
		err := ftry.sendTemplateOne(i)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ftry *RemoteFactory) sendTemplateOne(i int) error {
	cmd := exec.Command("scp",
		"-i", ftry.params.KeySSH,
		"-r", path.Join(ftry.templateDir, strconv.Itoa(i)),
		fmt.Sprintf("%s@%s:%s", ftry.params.LoginName, ftry.hosts[i],
			path.Join(ftry.params.RemoteWorkDir, "/template")),
	)
	return runCommand(cmd)
}

func (ftry *RemoteFactory) SetupCluster(name string) (*Cluster, error) {
	ftry.setupClusterDir(name)
	cls := &Cluster{
		nodes:      make([]Node, ftry.params.NodeCount),
		nodeConfig: ftry.params.NodeConfig,
	}
	cls.nodeConfig.Datadir = path.Join(ftry.params.RemoteWorkDir, name)
	juriaPath := path.Join(ftry.params.RemoteWorkDir, "juria")
	for i := 0; i < ftry.params.NodeCount; i++ {
		node := &RemoteNode{
			juriaPath: juriaPath,
			config:    cls.nodeConfig,
			loginName: ftry.params.LoginName,
			keySSH:    ftry.params.KeySSH,
			host:      ftry.hosts[i],
		}
		cls.nodes[i] = node
	}
	cls.Stop()
	return cls, nil
}

func (ftry *RemoteFactory) setupClusterDir(name string) {
	for i := 0; i < ftry.params.NodeCount; i++ {
		ftry.setupClusterDirOne(i, name)
	}
}

func (ftry *RemoteFactory) setupClusterDirOne(i int, name string) error {
	cmd := exec.Command("ssh",
		"-i", ftry.params.KeySSH,
		fmt.Sprintf("%s@%s", ftry.params.LoginName, ftry.hosts[i]),
		"cd", ftry.params.RemoteWorkDir, ";",
		"rm", "-r", name, ";",
		"cp", "-r", "template", name,
	)
	return cmd.Run()
}

type RemoteNode struct {
	juriaPath string
	config    node.Config

	loginName string
	keySSH    string
	host      string

	running bool
	mtxRun  sync.RWMutex
}

var _ Node = (*RemoteNode)(nil)

func (node *RemoteNode) Start() error {
	if node.IsRunning() {
		return nil
	}
	cmd := exec.Command("ssh",
		"-i", node.keySSH,
		fmt.Sprintf("%s@%s", node.loginName, node.host),
		"nohup", node.juriaPath,
		"-d", node.config.Datadir,
		"-p", strconv.Itoa(node.config.Port),
		"-P", strconv.Itoa(node.config.APIPort),
		"--debug", strconv.FormatBool(node.config.Debug),
		">>", path.Join(node.config.Datadir, "log.txt"), "2>&1", "&",
	)
	node.setRunning(true)
	return cmd.Run()
}

func (node *RemoteNode) Stop() {
	if !node.IsRunning() {
		return
	}
	node.setRunning(false)
	StopRemoteNode(node.host, node.loginName, node.keySSH)
}

func (node *RemoteNode) IsRunning() bool {
	node.mtxRun.RLock()
	defer node.mtxRun.RUnlock()
	return node.running
}

func (node *RemoteNode) setRunning(val bool) {
	node.mtxRun.Lock()
	defer node.mtxRun.Unlock()
	node.running = val
}

func (node *RemoteNode) GetEndpoint() string {
	return fmt.Sprintf("http://%s:%d", node.host, node.config.APIPort)
}

func StopRemoteNode(host, login, keySSH string) {
	cmd := exec.Command("ssh",
		"-i", keySSH,
		fmt.Sprintf("%s@%s", login, host),
		"sudo", "killall", "juria",
	)
	cmd.Run()
}
