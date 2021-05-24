// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package node

import (
	"crypto"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/aungmawjj/juria-blockchain/consensus"
	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/execution"
	"github.com/aungmawjj/juria-blockchain/logger"
	"github.com/aungmawjj/juria-blockchain/merkle"
	"github.com/aungmawjj/juria-blockchain/p2p"
	"github.com/aungmawjj/juria-blockchain/storage"
	"github.com/aungmawjj/juria-blockchain/txpool"
	"github.com/multiformats/go-multiaddr"
	_ "golang.org/x/crypto/sha3"
)

type Validator struct {
	PubKey []byte
	Addr   string
}

type Config struct {
	Debug   bool
	Datadir string
	Port    int
}

type Node struct {
	config Config

	privKey *core.PrivateKey

	vldKeys  []*core.PublicKey
	vldAddrs []multiaddr.Multiaddr

	vldStore  core.ValidatorStore
	storage   *storage.Storage
	host      *p2p.Host
	msgSvc    *p2p.MsgService
	txpool    *txpool.TxPool
	execution *execution.Execution
	consensus *consensus.Consensus
}

func Run(config Config) {
	node := new(Node)
	node.config = config
	node.setupLogger()
	if err := node.readKey(); err != nil {
		logger.Fatal("read key failed", "error", err)
	}
	if err := node.readValidators(); err != nil {
		logger.Fatal("read validators failed", "error", err)
	}

	node.vldStore = core.NewValidatorStore(node.vldKeys)
	if err := node.setupStorage(); err != nil {
		logger.Fatal("setup storage failed", "error", err)
	}
	if err := node.setupHost(); err != nil {
		logger.Fatal("setup p2p host failed", "error", err)
	}

	node.msgSvc = p2p.NewMsgService(node.host)
	node.txpool = txpool.New(node.storage, node.msgSvc)
	node.execution = execution.New(node.storage, execution.Config{})
	node.setupConsensus()
	node.msgSvc.SetReqHandler(&p2p.BlockReqHandler{
		GetBlock: node.GetBlock,
	})
	node.msgSvc.SetReqHandler(&p2p.TxListReqHandler{
		GetTxList: node.GetTxList,
	})
	select {}
}

func (node *Node) setupLogger() {
	var instance logger.Logger
	if node.config.Debug {
		instance = logger.NewWithConfig(logger.Config{
			Debug: true,
		})
	} else {
		instance = logger.New()
	}
	logger.Set(instance)
}

func (node *Node) readKey() error {
	b, err := ioutil.ReadFile(path.Join(node.config.Datadir, "nodekey"))
	if err != nil {
		return fmt.Errorf("cannot read nodekey %w", err)
	}
	node.privKey, err = core.NewPrivateKey(b)
	return err
}

func (node *Node) readValidators() error {
	f, err := os.Open(path.Join(node.config.Datadir, "validators.json"))
	if err != nil {
		return fmt.Errorf("cannot read validators %w", err)
	}
	defer f.Close()

	var vlds []Validator
	if err := json.NewDecoder(f).Decode(&vlds); err != nil {
		return fmt.Errorf("cannot parse validators json %w", err)
	}

	node.vldKeys = make([]*core.PublicKey, len(vlds))
	node.vldAddrs = make([]multiaddr.Multiaddr, len(vlds))

	for i, vld := range vlds {
		node.vldKeys[i], err = core.NewPublicKey(vld.PubKey)
		if err != nil {
			return fmt.Errorf("invalid public key %w", err)
		}
		node.vldAddrs[i], err = multiaddr.NewMultiaddr(vld.Addr)
		if err != nil {
			return fmt.Errorf("invalid multiaddr %w", err)
		}
	}
	return nil
}

func (node *Node) setupStorage() error {
	db, err := storage.NewDB(path.Join(node.config.Datadir, "db"))
	if err != nil {
		return fmt.Errorf("cannot create db %w", err)
	}
	node.storage = storage.New(db, merkle.TreeOptions{
		HashFunc:     crypto.SHA3_256,
		BranchFactor: 4,
	})
	return nil
}

func (node *Node) setupHost() error {
	addr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", node.config.Port))
	host, err := p2p.NewHost(node.privKey, addr)
	if err != nil {
		return fmt.Errorf("cannot create p2p host %w", err)
	}
	for i, key := range node.vldKeys {
		if !key.Equal(node.privKey.PublicKey()) {
			host.AddPeer(p2p.NewPeer(key, node.vldAddrs[i]))
		}
	}
	node.host = host
	return nil
}

func (node *Node) setupConsensus() {
	node.consensus = consensus.New(&consensus.Resources{
		Signer:    node.privKey,
		VldStore:  node.vldStore,
		Storage:   node.storage,
		MsgSvc:    node.msgSvc,
		TxPool:    node.txpool,
		Execution: node.execution,
	}, consensus.Config{})

}

func (node *Node) GetBlock(sender *core.PublicKey, hash []byte) (*core.Block, error) {
	if blk := node.consensus.GetBlock(hash); blk != nil {
		return blk, nil
	}
	return node.storage.GetBlock(hash)
}

func (node *Node) GetTxList(sender *core.PublicKey, hashes [][]byte) (*core.TxList, error) {
	ret := make(core.TxList, len(hashes))
	for i, hash := range hashes {
		tx := node.txpool.GetTx(hash)
		if tx != nil {
			ret[i] = tx
			continue
		}
		tx, err := node.storage.GetTx(hash)
		if err != nil {
			return nil, err
		}
		ret[i] = tx
	}
	return &ret, nil
}
