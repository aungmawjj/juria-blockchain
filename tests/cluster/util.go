// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package cluster

import (
	"encoding/json"
	"os"
	"path"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/node"
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
