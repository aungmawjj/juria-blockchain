// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package cluster

import (
	"fmt"
	"os"
	"os/exec"
	"path"
)

type localNode struct {
	juriaPath string

	datadir string
	port    string
	apiPort string

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
	cmd := exec.Command(node.juriaPath,
		"-d", node.datadir,
		"-p", node.port,
	)
	cmd.Stderr = node.logFile
	cmd.Stdout = node.logFile
	return cmd.Start()
}

func (node *localNode) Stop() {
	if !node.running {
		return
	}
	node.cmd.Process.Kill()
	node.logFile.Close()
}

func (node *localNode) GetEndpoint() string {
	return fmt.Sprintf("http://172.0.0.1:%s", node.apiPort)
}
