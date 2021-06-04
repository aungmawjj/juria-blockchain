// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package node

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/aungmawjj/juria-blockchain/consensus"
	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/execution"
	"github.com/aungmawjj/juria-blockchain/logger"
	"github.com/aungmawjj/juria-blockchain/p2p"
	"github.com/aungmawjj/juria-blockchain/storage"
	"github.com/aungmawjj/juria-blockchain/txpool"
	"github.com/multiformats/go-multiaddr"
	"go.uber.org/zap"
)

type Node struct {
	config Config

	privKey *core.PrivateKey
	peers   []*p2p.Peer
	genesis *Genesis

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
	node.setupBinccDir()
	node.setupLogger()
	node.readFiles()
	node.setupComponents()
	logger.I().Infow("node setup done, starting consensus...")
	node.consensus.Start()
	status := node.consensus.GetStatus()
	logger.I().Infow("started consensus",
		"leader", status.LeaderIndex, "bLeaf", status.BLeaf, "qc", status.QCHigh)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	logger.I().Info("node killed")
	node.consensus.Stop()
}

func (node *Node) setupLogger() {
	var inst *zap.Logger
	var err error
	if node.config.Debug {
		inst, err = zap.NewDevelopment()
	} else {
		inst, err = zap.NewProduction()
	}
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	logger.Set(inst.Sugar())
}

func (node *Node) setupBinccDir() {
	node.config.ExecutionConfig.BinccDir = path.Join(node.config.Datadir, "bincc")
	os.Mkdir(node.config.ExecutionConfig.BinccDir, 0755)
}

func (node *Node) readFiles() {
	var err error
	node.privKey, err = readNodeKey(node.config.Datadir)
	if err != nil {
		logger.I().Fatalw("read key failed", "error", err)
	}
	logger.I().Infow("read nodekey", "pubkey", node.privKey.PublicKey())

	node.genesis, err = readGenesis(node.config.Datadir)
	if err != nil {
		logger.I().Fatalw("read genesis failed", "error", err)
	}

	node.peers, err = readPeers(node.config.Datadir)
	if err != nil {
		logger.I().Fatalw("read peers failed", "error", err)
	}
	logger.I().Infow("read peers", "count", len(node.peers))
}

func (node *Node) setupComponents() {
	node.setupValidatorStore()
	node.setupStorage()
	node.setupHost()
	logger.I().Infow("setup p2p host", "port", node.config.Port)
	node.msgSvc = p2p.NewMsgService(node.host)
	node.execution = execution.New(node.storage, node.config.ExecutionConfig)
	node.txpool = txpool.New(node.storage, node.execution, node.msgSvc)
	node.setupConsensus()
	node.setReqHandlers()
	serveNodeAPI(node)
}

func (node *Node) setupValidatorStore() {
	validators := make([]*core.PublicKey, len(node.genesis.Validators))
	for i, v := range node.genesis.Validators {
		pubKey, err := core.NewPublicKey(v)
		if err != nil {
			logger.I().Fatalw("parse validator failed", "error", err)
		}
		validators[i] = pubKey
	}
	node.vldStore = core.NewValidatorStore(validators)
}

func (node *Node) setupStorage() {
	db, err := storage.NewDB(path.Join(node.config.Datadir, "db"))
	if err != nil {
		logger.I().Fatalw("setup storage failed", "error", err)
	}
	node.storage = storage.New(db, node.config.StorageConfig)
}

func (node *Node) setupHost() {
	ln, err := net.Listen("tcp4", fmt.Sprintf(":%d", node.config.Port))
	if err != nil {
		logger.I().Fatalw("cannot listen on port", "port", node.config.Port)
	}
	ln.Close()
	addr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", node.config.Port))
	host, err := p2p.NewHost(node.privKey, addr)
	if err != nil {
		logger.I().Fatalw("cannot create p2p host", "error", err)
	}
	for _, p := range node.peers {
		if !p.PublicKey().Equal(node.privKey.PublicKey()) {
			host.AddPeer(p)
		}
	}
	node.host = host
}

func (node *Node) setupConsensus() {
	node.consensus = consensus.New(&consensus.Resources{
		Signer:    node.privKey,
		VldStore:  node.vldStore,
		Storage:   node.storage,
		MsgSvc:    node.msgSvc,
		TxPool:    node.txpool,
		Execution: node.execution,
	}, node.config.ConsensusConfig)

}

func (node *Node) setReqHandlers() {
	node.msgSvc.SetReqHandler(&p2p.BlockReqHandler{
		GetBlock: node.GetBlock,
	})
	node.msgSvc.SetReqHandler(&p2p.BlockByHeightReqHandler{
		GetBlockByHeight: node.storage.GetBlockByHeight,
	})
	node.msgSvc.SetReqHandler(&p2p.TxListReqHandler{
		GetTxList: node.GetTxList,
	})
}

func (node *Node) GetBlock(hash []byte) (*core.Block, error) {
	if blk := node.consensus.GetBlock(hash); blk != nil {
		return blk, nil
	}
	return node.storage.GetBlock(hash)
}

func (node *Node) GetTxList(hashes [][]byte) (*core.TxList, error) {
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
