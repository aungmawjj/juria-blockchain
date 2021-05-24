// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package main

import (
	"log"

	"github.com/aungmawjj/juria-blockchain/node"
	"github.com/spf13/cobra"
)

const (
	flagDebug   = "debug"
	flagDataDir = "datadir"
	flagPort    = "port"
)

var rootCmd = &cobra.Command{
	Use:   "juria",
	Short: "Juria blockchain",
	Run: func(cmd *cobra.Command, args []string) {
		debug, err := cmd.Flags().GetBool(flagDebug)
		check(err)
		datadir, err := cmd.Flags().GetString(flagDataDir)
		check(err)
		port, err := cmd.Flags().GetInt(flagPort)
		check(err)

		node.Run(node.Config{
			Debug:   debug,
			Datadir: datadir,
			Port:    port,
		})
	},
}

func main() {
	check(rootCmd.Execute())
}

func init() {
	rootCmd.PersistentFlags().Bool(flagDebug, false, "debug mode")
	rootCmd.PersistentFlags().StringP(flagDataDir, "d", "", "blockchain data directory")
	rootCmd.MarkPersistentFlagRequired(flagDataDir)

	rootCmd.Flags().IntP(flagPort, "p", 9040, "p2p port")
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
