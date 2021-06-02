// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package cluster

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/node"
	"github.com/multiformats/go-multiaddr"
)

func GetRemoteHosts(hostsPath string, nodeCount int) ([]string, error) {
	raw, err := ioutil.ReadFile(hostsPath)
	if err != nil {
		return nil, err
	}
	hosts := strings.Split(string(raw), "\n")
	if len(hosts) < nodeCount {
		return nil, fmt.Errorf("not enough hosts, %d | %d",
			len(hosts), nodeCount)
	}
	return hosts, nil
}

func StopRemoteNode(host, login, keySSH string) {
	cmd := exec.Command("ssh",
		"-i", keySSH,
		fmt.Sprintf("%s@%s", login, host),
		"sudo", "killall", "juria",
	)
	cmd.Run()
}

func WriteNodeKey(nodedir string, key *core.PrivateKey) error {
	f, err := os.Create(path.Join(nodedir, node.NodekeyFile))
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(key.Bytes())
	return err
}

func WriteValidatorsFile(nodedir string, vlds []node.Validator) error {
	f, err := os.Create(path.Join(nodedir, node.ValidatorsFile))
	if err != nil {
		return err
	}
	defer f.Close()
	e := json.NewEncoder(f)
	e.SetIndent("", "  ")
	return e.Encode(vlds)
}

func MakeRandomKeys(count int) []*core.PrivateKey {
	keys := make([]*core.PrivateKey, count)
	for i := 0; i < count; i++ {
		keys[i] = core.GenerateKey(nil)
	}
	return keys
}

func MakeValidators(keys []*core.PrivateKey, addrs []multiaddr.Multiaddr) []node.Validator {
	vlds := make([]node.Validator, len(addrs))
	// create validator infos (pubkey + addr)
	for i, addr := range addrs {
		vlds[i] = node.Validator{
			PubKey: keys[i].PublicKey().Bytes(),
			Addr:   addr.String(),
		}
	}
	return vlds
}

func SetupTemplateDir(dir string, keys []*core.PrivateKey, vlds []node.Validator) error {
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	if err := os.Mkdir(dir, 0755); err != nil {
		return err
	}
	for i, key := range keys {
		dir := path.Join(dir, strconv.Itoa(i))
		os.Mkdir(dir, 0755)
		if err := WriteNodeKey(dir, key); err != nil {
			return err
		}
		if err := WriteValidatorsFile(dir, vlds); err != nil {
			return err
		}
	}
	return nil
}

func RunCommand(cmd *exec.Cmd) error {
	cmd.Stdout = os.Stdout
	fmt.Printf(" $ %s\n", strings.Join(cmd.Args, " "))
	return cmd.Run()
}

func AddJuriaFlags(cmd *exec.Cmd, config *node.Config) {
	cmd.Args = append(cmd.Args, "-d", config.Datadir)
	cmd.Args = append(cmd.Args, "-p", strconv.Itoa(config.Port))
	cmd.Args = append(cmd.Args, "-P", strconv.Itoa(config.APIPort))
	cmd.Args = append(cmd.Args, "--debug", strconv.FormatBool(config.Debug))

	cmd.Args = append(cmd.Args, "--storage-merkleBranchFactor",
		strconv.Itoa(int(config.StorageConfig.MerkleBranchFactor)))

	cmd.Args = append(cmd.Args, "--execution-txExecTimeout",
		config.ExecutionConfig.TxExecTimeout.String(),
	)
	cmd.Args = append(cmd.Args, "--execution-concurrentLimit",
		strconv.Itoa(config.ExecutionConfig.ConcurrentLimit))

	cmd.Args = append(cmd.Args, "--chainid",
		strconv.Itoa(int(config.ConsensusConfig.ChainID)))

	cmd.Args = append(cmd.Args, "--consensus-blockTxLimit",
		strconv.Itoa(config.ConsensusConfig.BlockTxLimit))

	cmd.Args = append(cmd.Args, "--consensus-txWaitTime",
		config.ConsensusConfig.TxWaitTime.String())

	cmd.Args = append(cmd.Args, "--consensus-beatTimeout",
		config.ConsensusConfig.BeatTimeout.String())

	cmd.Args = append(cmd.Args, "--consensus-blockDelay",
		config.ConsensusConfig.BlockDelay.String())

	cmd.Args = append(cmd.Args, "--consensus-viewWidth",
		config.ConsensusConfig.ViewWidth.String())

	cmd.Args = append(cmd.Args, "--consensus-leaderTimeout",
		config.ConsensusConfig.LeaderTimeout.String())
}
