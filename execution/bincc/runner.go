// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package bincc

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"github.com/aungmawjj/juria-blockchain/execution/chaincode"
	"github.com/aungmawjj/juria-blockchain/logger"
)

const MessageSizeLimit = 100 * 1000 * 1000

type Runner struct {
	codePath string
	timeout  time.Duration

	callContext chaincode.CallContext

	cmd   *exec.Cmd
	rw    *readWriter
	timer *time.Timer
}

var _ chaincode.Chaincode = (*Runner)(nil)

func (r *Runner) Init(ctx chaincode.CallContext) error {
	r.callContext = ctx
	_, err := r.runCode(CallTypeInit)
	return err
}

func (r *Runner) Invoke(ctx chaincode.CallContext) error {
	r.callContext = ctx
	_, err := r.runCode(CallTypeInvoke)
	return err
}

func (r *Runner) Query(ctx chaincode.CallContext) ([]byte, error) {
	r.callContext = ctx
	return r.runCode(CallTypeQuery)
}

func (r *Runner) runCode(callType CallType) ([]byte, error) {
	r.timer = time.NewTimer(r.timeout)
	defer r.timer.Stop()

	if err := r.startCode(CallTypeInit); err != nil {
		return nil, err
	}
	defer r.cmd.Process.Kill()

	if err := r.sendCallData(callType); err != nil {
		return nil, err
	}
	res, err := r.serveStateAndGetResult()
	if err != nil {
		return nil, err
	}
	return res, r.cmd.Wait()
}

func (r *Runner) startCode(callType CallType) error {
	if err := r.setupCmd(); err != nil {
		return err
	}
	err := r.cmd.Start()
	if err == nil {
		return nil
	}
	logger.I().Warnf("start code error %f", err)
	select {
	case <-r.timer.C:
		return fmt.Errorf("chaincode start timeout")
	default:
	}
	time.Sleep(5 * time.Millisecond)
	return r.startCode(callType)
}

func (r *Runner) setupCmd() error {
	r.cmd = exec.Command(r.codePath)
	var err error
	r.rw = new(readWriter)
	r.rw.writer, err = r.cmd.StdinPipe()
	if err != nil {
		return err
	}
	r.rw.reader, err = r.cmd.StderrPipe()
	return err
}

func (r *Runner) sendCallData(callType CallType) error {
	callData := &CallData{
		CallType:    callType,
		Input:       r.callContext.Input(),
		Sender:      r.callContext.Sender(),
		BlockHash:   r.callContext.BlockHash(),
		BlockHeight: r.callContext.BlockHeight(),
	}
	b, _ := json.Marshal(callData)
	return r.rw.write(b)
}

func (r *Runner) serveStateAndGetResult() ([]byte, error) {
	for {
		select {
		case <-r.timer.C:
			return nil, fmt.Errorf("chaincode call timeout")
		default:
		}
		b, err := r.rw.read()
		if err != nil {
			return nil, fmt.Errorf("read upstream error %w", err)
		}
		up := new(UpStream)
		if err := json.Unmarshal(b, up); err != nil {
			return nil, fmt.Errorf("cannot parse upstream data")
		}
		if up.Type == UpStreamResult {
			if len(up.Error) > 0 {
				return nil, fmt.Errorf(up.Error)
			}
			return up.Value, nil
		}
		if err := r.serveState(up); err != nil {
			return nil, err
		}
	}
}

func (r *Runner) serveState(up *UpStream) error {
	down := new(DownStream)
	switch up.Type {

	case UpStreamGetState:
		val := r.callContext.GetState(up.Key)
		down.Value = val

	case UpStreamVerifyState:
		val, err := r.callContext.VerifyState(up.Key)
		down.Value = val
		if err != nil {
			down.Error = err.Error()
		}

	case UpStreamSetState:
		r.callContext.SetState(up.Key, up.Value)
	}

	b, _ := json.Marshal(down)
	return r.rw.write(b)
}
