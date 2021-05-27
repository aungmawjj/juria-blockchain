// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package node

import (
	"fmt"
	"io"
	"net/http"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/logger"
	"github.com/gin-gonic/gin"
)

type RequestByHash struct {
	Hash []byte `json:"hash"`
}

type nodeAPI struct {
	node *Node
}

func serveNodeAPI(node *Node) {
	api := &nodeAPI{node}

	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/consensus", api.getConsensusStatus)

	r.GET("/txpool", api.getTxPoolStatus)
	r.POST("/transaction", api.submitTX)
	r.GET("/transaction/commit", api.getTxCommit)

	r.GET("/block", api.getBlock)

	go func() {
		err := r.Run(fmt.Sprintf(":%d", node.config.APIPort))
		if err != nil {
			logger.I().Fatalf("failed to start api %+v", err)
		}
	}()
}

func (api *nodeAPI) getConsensusStatus(c *gin.Context) {
	fmt.Println(api.node.consensus.GetStatus())
	c.JSON(http.StatusOK, api.node.consensus.GetStatus())
}

func (api *nodeAPI) getTxPoolStatus(c *gin.Context) {
	c.JSON(http.StatusOK, api.node.txpool.GetStatus())
}

func (api *nodeAPI) submitTX(c *gin.Context) {
	tx := core.NewTransaction()
	if err := c.ShouldBind(tx); err != nil {
		c.String(http.StatusBadRequest, "cannot parse input")
		return
	}
	if err := api.node.txpool.SubmitTx(tx); err != nil {
		logger.I().Warnf("submit tx failed %+v", err)
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.String(http.StatusOK, "transaction accepted")
}

func (api *nodeAPI) getTxCommit(c *gin.Context) {
	req := new(RequestByHash)
	if err := c.ShouldBindQuery(req); err != nil {
		c.String(http.StatusBadRequest, "cannot parse request")
		return
	}
	txc, err := api.node.storage.GetTxCommit(req.Hash)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, txc)
}

func (api *nodeAPI) getBlock(c *gin.Context) {
	req := new(RequestByHash)
	if err := c.ShouldBindQuery(req); err != nil {
		c.String(http.StatusBadRequest, "cannot parse request")
		return
	}
	blk, err := api.node.GetBlock(req.Hash)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, blk)
}
