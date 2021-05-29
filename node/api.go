// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package node

import (
	"encoding/hex"
	"fmt"
	"io"
	"net/http"

	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/logger"
	"github.com/gin-gonic/gin"
)

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
	r.POST("/transactions", api.submitTX)
	r.GET("/transactions/:hash/status", api.getTxStatus)
	r.GET("/transactions/:hash/commit", api.getTxCommit)

	r.GET("/blocks/:hash", api.getBlock)
	r.GET("/blocksbyh/:height", api.getBlockByHeight)

	r.POST("/querystate", api.queryState)

	go func() {
		err := r.Run(fmt.Sprintf(":%d", node.config.APIPort))
		if err != nil {
			logger.I().Fatalf("failed to start api %+v", err)
		}
	}()
}

func (api *nodeAPI) getConsensusStatus(c *gin.Context) {
	c.JSON(http.StatusOK, api.node.consensus.GetStatus())
}

func (api *nodeAPI) getTxPoolStatus(c *gin.Context) {
	c.JSON(http.StatusOK, api.node.txpool.GetStatus())
}

func (api *nodeAPI) submitTX(c *gin.Context) {
	tx := core.NewTransaction()
	if err := c.ShouldBind(tx); err != nil {
		c.String(http.StatusBadRequest, "cannot parse tx")
		return
	}
	if err := api.node.txpool.SubmitTx(tx); err != nil {
		logger.I().Warnf("submit tx failed %+v", err)
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.String(http.StatusOK, "transaction accepted")
}

type StateQuery struct {
	CodeAddr []byte
	Input    []byte
}

func (api *nodeAPI) queryState(c *gin.Context) {
	query := new(StateQuery)
	if err := c.ShouldBind(query); err != nil {
		c.String(http.StatusBadRequest, "cannot parse request")
		return
	}
	result, err := api.node.execution.Query(query.CodeAddr, query.Input)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

func (api *nodeAPI) getTxStatus(c *gin.Context) {
	hash, err := api.getHash(c)
	if err != nil {
		c.String(http.StatusBadRequest, "cannot parse request")
		return
	}
	status := api.node.txpool.GetTxStatus(hash)
	c.JSON(http.StatusOK, status)
}

func (api *nodeAPI) getTxCommit(c *gin.Context) {
	hash, err := api.getHash(c)
	if err != nil {
		c.String(http.StatusBadRequest, "cannot parse request")
		return
	}
	txc, err := api.node.storage.GetTxCommit(hash)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, txc)
}

func (api *nodeAPI) getBlock(c *gin.Context) {
	hash, err := api.getHash(c)
	if err != nil {
		c.String(http.StatusBadRequest, "cannot parse request")
		return
	}
	blk, err := api.node.GetBlock(hash)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, blk)
}

func (api *nodeAPI) getHash(c *gin.Context) ([]byte, error) {
	hashstr := c.Param("hash")
	return hex.DecodeString(hashstr)
}

func (api *nodeAPI) getBlockByHeight(c *gin.Context) {
	height := c.GetUint64("height")
	blk, err := api.node.storage.GetBlockByHeight(height)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, blk)
}
