// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package bincc

import (
	"fmt"
	"os"
	"time"

	"github.com/aungmawjj/juria-blockchain/execution/bincc/bincc_pb"
	"github.com/aungmawjj/juria-blockchain/execution/chaincode"
	"google.golang.org/protobuf/proto"
)

const ChaincodeHardTimeout = 10 * time.Second

type Client struct {
	rw       *readWriter
	cc       chaincode.Chaincode
	callData *bincc_pb.CallData
}

var _ chaincode.CallContext = (*Client)(nil)

func RunChaincode(cc chaincode.Chaincode) {
	timeout := time.After(ChaincodeHardTimeout)
	done := make(chan struct{})
	go runChaincodeAsync(cc, done)
	select {
	case <-timeout:
		os.Exit(1)
	case <-done:
		os.Exit(0)
	}
}

func runChaincodeAsync(cc chaincode.Chaincode, done chan<- struct{}) {
	defer close(done)
	c := &Client{
		rw: &readWriter{
			reader: os.Stdin,
			writer: os.Stderr,
		},
		cc: cc,
	}
	if err := c.loadCallData(); err != nil {
		return
	}
	c.runChaincode()
}

func (c *Client) loadCallData() error {
	b, err := c.rw.read()
	if err != nil {
		return err
	}
	c.callData = new(bincc_pb.CallData)
	return proto.Unmarshal(b, c.callData)
}

func (c *Client) runChaincode() {
	var result []byte
	var err error

	switch c.callData.CallType {
	case bincc_pb.CallType_Init:
		err = c.cc.Init(c)

	case bincc_pb.CallType_Invoke:
		err = c.cc.Invoke(c)

	case bincc_pb.CallType_Query:
		result, err = c.cc.Query(c)
	}
	c.sendResult(result, err)
}

func (c *Client) Sender() []byte {
	return c.callData.Sender
}

func (c *Client) BlockHash() []byte {
	return c.callData.BlockHash
}

func (c *Client) BlockHeight() uint64 {
	return c.callData.BlockHeight
}

func (c *Client) Input() []byte {
	return c.callData.Input
}

func (c *Client) VerifyState(key []byte) ([]byte, error) {
	return c.request(key, nil, bincc_pb.UpStream_VerifyState)
}

func (c *Client) GetState(key []byte) []byte {
	val, _ := c.request(key, nil, bincc_pb.UpStream_GetState)
	return val
}

func (c *Client) SetState(key, value []byte) {
	c.request(key, value, bincc_pb.UpStream_SetState)
}

func (c *Client) request(key, value []byte, upType bincc_pb.UpStream_Type) ([]byte, error) {
	up := new(bincc_pb.UpStream)
	up.Type = upType
	up.Key = key
	up.Value = value
	b, _ := proto.Marshal(up)
	if err := c.rw.write(b); err != nil {
		return nil, err
	}
	b, err := c.rw.read()
	if err != nil {
		return nil, err
	}
	down := new(bincc_pb.DownStream)
	if err := proto.Unmarshal(b, down); err != nil {
		return nil, err
	}
	if len(down.Error) > 0 {
		return nil, fmt.Errorf(down.Error)
	}
	return down.Value, nil
}

func (c *Client) sendResult(value []byte, err error) {
	up := new(bincc_pb.UpStream)
	up.Type = bincc_pb.UpStream_Result
	up.Value = value
	if err != nil {
		up.Error = err.Error()
	}
	b, _ := proto.Marshal(up)
	c.rw.write(b)
}
