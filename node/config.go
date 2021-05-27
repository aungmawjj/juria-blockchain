// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package node

import (
	"github.com/aungmawjj/juria-blockchain/consensus"
	"github.com/aungmawjj/juria-blockchain/execution"
	"github.com/aungmawjj/juria-blockchain/storage"
)

type Config struct {
	Debug   bool
	Datadir string
	Port    int
	APIPort int

	StorageConfig   storage.Config
	ExecutionConfig execution.Config
	ConsensusConfig consensus.Config
}

var DefaultConfig = Config{
	Port:            15150,
	APIPort:         9040,
	StorageConfig:   storage.DefaultConfig,
	ExecutionConfig: execution.DefaultConfig,
	ConsensusConfig: consensus.DefaultConfig,
}
