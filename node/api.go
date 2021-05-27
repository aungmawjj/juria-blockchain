// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package node

import (
	"encoding/base64"
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
	r.POST("/transactions/:hash", api.submitTX)
	r.GET("/transactions/:hash/commit", api.getTxCommit)

	r.GET("/blocks/:hash", api.getBlock)
	r.GET("/blocksbyh/:height", api.getBlockByHeight)

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
	return base64.StdEncoding.DecodeString(hashstr)
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
