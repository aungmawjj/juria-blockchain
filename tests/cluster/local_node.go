// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package cluster

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
)

type localNode struct {
	juriaPath string

	datadir string
	port    int
	apiPort int

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
	node.cmd.Process.Kill()
	node.logFile.Close()
}

func (node *localNode) GetEndpoint() string {
	return fmt.Sprintf("http://127.0.0.1:%d", node.apiPort)
}
