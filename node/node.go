// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package node

import (
	"fmt"
	"log"
	"net"
	"path"

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

	privKey  *core.PrivateKey
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
	node.readFiles()
	node.setupComponents()
	logger.I().Infow("node setup done, starting consensus...")
	node.consensus.Start()
	status := node.consensus.GetStatus()
	logger.I().Infow("started consensus",
		"leader", status.LeaderIndex, "bLeaf", status.BLeaf, "qc", status.QCHigh)
	select {}
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

func (node *Node) readFiles() {
	var err error
	node.privKey, err = readNodeKey(node.config.Datadir)
	if err != nil {
		logger.I().Fatalw("read key failed", "error", err)
	}
	logger.I().Infow("read nodekey", "pubkey", node.privKey.PublicKey())

	node.vldKeys, node.vldAddrs, err = readValidators(node.config.Datadir)
	if err != nil {
		logger.I().Fatalw("read validators failed", "error", err)
	}
	logger.I().Infow("read validators", "count", len(node.vldKeys))
}

func (node *Node) setupComponents() {
	node.vldStore = core.NewValidatorStore(node.vldKeys)
	if err := node.setupStorage(); err != nil {
		logger.I().Fatalw("setup storage failed", "error", err)
	}
	if err := node.setupHost(); err != nil {
		logger.I().Fatalw("setup p2p host failed", "error", err)
	}
	logger.I().Infow("setup p2p host", "port", node.config.Port)
	node.msgSvc = p2p.NewMsgService(node.host)
	node.txpool = txpool.New(node.storage, node.msgSvc)
	node.execution = execution.New(node.storage, node.config.ExecutionConfig)
	node.setupConsensus()
	node.msgSvc.SetReqHandler(&p2p.BlockReqHandler{
		GetBlock: node.GetBlock,
	})
	node.msgSvc.SetReqHandler(&p2p.BlockByHeightReqHandler{
		GetBlockByHeight: node.storage.GetBlockByHeight,
	})
	node.msgSvc.SetReqHandler(&p2p.TxListReqHandler{
		GetTxList: node.GetTxList,
	})
	serveNodeAPI(node)
}

func (node *Node) setupStorage() error {
	db, err := storage.NewDB(path.Join(node.config.Datadir, "db"))
	if err != nil {
		return fmt.Errorf("cannot create db %w", err)
	}
	node.storage = storage.New(db, node.config.StorageConfig)
	return nil
}

func (node *Node) setupHost() error {
	ln, err := net.Listen("tcp4", fmt.Sprintf(":%d", node.config.Port))
	if err != nil {
		return fmt.Errorf("cannot listen on %d, %w", node.config.Port, err)
	}
	ln.Close()
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
	}, node.config.ConsensusConfig)

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
