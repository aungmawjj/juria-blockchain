// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strings"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/node"
	"github.com/multiformats/go-multiaddr"
)

const (
	Juria     = "./juria"
	TestDir   = "test-data"
	NodeCount = 4
)

func main() {
	runNodes(Juria, TestDir, NodeCount)
}

func runNodes(juria, rootdir string, count int) {
	err := os.RemoveAll(rootdir)
	check(err)
	err = os.Mkdir(rootdir, 0755)
	check(err)

	keys := make([]*core.PrivateKey, count)
	ports := make([]string, count)
	addrs := make([]multiaddr.Multiaddr, count)
	vlds := make([]node.Validator, count)
	dirs := make([]string, count)

	for i := 0; i < count; i++ {
		keys[i] = core.GenerateKey(nil)
		ports[i] = fmt.Sprintf("%d", 9040+i)
		addr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%s", ports[i]))
		addrs[i] = addr
		vlds[i] = node.Validator{
			PubKey: keys[i].PublicKey().Bytes(),
			Addr:   addrs[i].String(),
		}
		dirs[i] = path.Join(rootdir, fmt.Sprintf("%0d", i))
	}

	for i := 0; i < count; i++ {
		os.Mkdir(dirs[i], 0755)
		f, _ := os.Create(path.Join(dirs[i], "nodekey"))
		f.Write(keys[i].Bytes())
		f.Close()

		f, _ = os.Create(path.Join(dirs[i], "validators.json"))
		e := json.NewEncoder(f)
		e.SetIndent("", "  ")
		e.Encode(vlds)
	}

	cmds := make([]*exec.Cmd, count)
	for i := 0; i < count; i++ {
		logfile, _ := os.Create(path.Join(dirs[i], "log.txt"))
		cmd := exec.Command(juria, "-d", dirs[i], "-p", ports[i])
		fmt.Println(strings.Join(cmd.Args, " "))
		cmd.Stderr = logfile
		cmd.Stdout = logfile
		cmd.Start()
		cmds[i] = cmd
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Block until a signal is received.
	s := <-c
	fmt.Println("Got signal:", s)
	for _, cmd := range cmds {
		cmd.Process.Kill()
	}
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
