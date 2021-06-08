// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package main

import (
	"github.com/aungmawjj/juria-blockchain/chaincodes/juriacoin"
	"github.com/aungmawjj/juria-blockchain/execution/bincc"
)

// bincc version of juriacoin. User can compile and deploy it separately to the running juria network

func main() {
	jcc := new(juriacoin.JuriaCoin)
	bincc.RunChaincode(jcc)
}
