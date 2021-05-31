// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package cluster

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/node"
	"github.com/multiformats/go-multiaddr"
)

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

func makeRandomKeys(count int) []*core.PrivateKey {
	keys := make([]*core.PrivateKey, count)
	for i := 0; i < count; i++ {
		keys[i] = core.GenerateKey(nil)
	}
	return keys
}

func makeValidators(keys []*core.PrivateKey, addrs []multiaddr.Multiaddr) []node.Validator {
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

func setupTemplateDir(dir string, keys []*core.PrivateKey, vlds []node.Validator) error {
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	if err := os.Mkdir(dir, 0755); err != nil {
		return err
	}
	for i, key := range keys {
		dir := path.Join(dir, strconv.Itoa(i))
		os.Mkdir(dir, 0755)
		if err := writeNodeKey(dir, key); err != nil {
			return err
		}
		if err := writeValidatorsFile(dir, vlds); err != nil {
			return err
		}
	}
	return nil
}

func runCommand(cmd *exec.Cmd) error {
	cmd.Stdout = os.Stdout
	fmt.Printf(" $ %s\n", strings.Join(cmd.Args, " "))
	return cmd.Run()
}
