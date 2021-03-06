// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package main

import (
	"log"

	"github.com/aungmawjj/juria-blockchain/node"
	"github.com/spf13/cobra"
)

const (
	FlagDebug   = "debug"
	FlagDataDir = "datadir"

	FlagPort    = "port"
	FlagAPIPort = "apiPort"

	// storage
	FlagMerkleBranchFactor = "storage-merkleBranchFactor"

	// execution
	FlagTxExecTimeout       = "execution-txExecTimeout"
	FlagExecConcurrentLimit = "execution-concurrentLimit"

	// consensus
	FlagChainID       = "chainid"
	FlagBlockTxLimit  = "consensus-blockTxLimit"
	FlagTxWaitTime    = "consensus-txWaitTime"
	FlagBeatTimeout   = "consensus-beatTimeout"
	FlagBlockDelay    = "consensus-blockDelay"
	FlagViewWidth     = "consensus-viewWidth"
	FlagLeaderTimeout = "consensus-leaderTimeout"
)

var nodeConfig = node.DefaultConfig

var rootCmd = &cobra.Command{
	Use:   "juria",
	Short: "Juria blockchain",
	Run: func(cmd *cobra.Command, args []string) {
		node.Run(nodeConfig)
	},
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&nodeConfig.Debug,
		FlagDebug, false, "debug mode")

	rootCmd.PersistentFlags().StringVarP(&nodeConfig.Datadir,
		FlagDataDir, "d", "", "blockchain data directory")
	rootCmd.MarkPersistentFlagRequired(FlagDataDir)

	rootCmd.Flags().IntVarP(&nodeConfig.Port,
		FlagPort, "p", nodeConfig.Port, "p2p port")

	rootCmd.Flags().IntVarP(&nodeConfig.APIPort,
		FlagAPIPort, "P", nodeConfig.APIPort, "node api port")

	rootCmd.Flags().Uint8Var(&nodeConfig.StorageConfig.MerkleBranchFactor,
		FlagMerkleBranchFactor, nodeConfig.StorageConfig.MerkleBranchFactor,
		"merkle tree branching factor")

	rootCmd.Flags().DurationVar(&nodeConfig.ExecutionConfig.TxExecTimeout,
		FlagTxExecTimeout, nodeConfig.ExecutionConfig.TxExecTimeout,
		"tx execution timeout")

	rootCmd.Flags().IntVar(&nodeConfig.ExecutionConfig.ConcurrentLimit,
		FlagExecConcurrentLimit, nodeConfig.ExecutionConfig.ConcurrentLimit,
		"concurrent tx execution limit")

	rootCmd.Flags().Int64Var(&nodeConfig.ConsensusConfig.ChainID,
		FlagChainID, nodeConfig.ConsensusConfig.ChainID,
		"chainid is used to create genesis block")

	rootCmd.Flags().IntVar(&nodeConfig.ConsensusConfig.BlockTxLimit,
		FlagBlockTxLimit, nodeConfig.ConsensusConfig.BlockTxLimit,
		"maximum tx count in a block")

	rootCmd.Flags().DurationVar(&nodeConfig.ConsensusConfig.TxWaitTime,
		FlagTxWaitTime, nodeConfig.ConsensusConfig.TxWaitTime,
		"block creation delay if no transactions in the pool")

	rootCmd.Flags().DurationVar(&nodeConfig.ConsensusConfig.BeatTimeout,
		FlagBeatTimeout, nodeConfig.ConsensusConfig.BeatTimeout,
		"duration to wait to propose next block if leader cannot create qc")

	rootCmd.Flags().DurationVar(&nodeConfig.ConsensusConfig.BlockDelay,
		FlagBlockDelay, nodeConfig.ConsensusConfig.BlockDelay,
		"minimum delay between blocks")

	rootCmd.Flags().DurationVar(&nodeConfig.ConsensusConfig.ViewWidth,
		FlagViewWidth, nodeConfig.ConsensusConfig.ViewWidth,
		"view duration for a leader")

	rootCmd.Flags().DurationVar(&nodeConfig.ConsensusConfig.LeaderTimeout,
		FlagLeaderTimeout, nodeConfig.ConsensusConfig.LeaderTimeout,
		"leader must create next qc in this duration")
}
