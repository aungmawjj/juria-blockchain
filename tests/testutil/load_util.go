// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package testutil

import (
	"bytes"
	"net/http"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/tests/cluster"
)

func submitTxAndWait(cls *cluster.Cluster, tx *core.Transaction) error {
	if err := submitTx(cls, tx); err != nil {
		return err
	}
	return nil
}

func submitTx(cls *cluster.Cluster, tx *core.Transaction) error {
	b, err := tx.Marshal()
	if err != nil {
		return err
	}
	for i := 0; i < cls.NodeCount(); i++ {
		resp, err := http.Post(cls.GetNode(i).GetEndpoint()+"/transactions",
			"application-json", bytes.NewReader(b))
		err = checkResponse(resp, err)
		if err == nil {
			return nil
		}
	}
	return nil
}
