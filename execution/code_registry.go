// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package execution

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/aungmawjj/juria-blockchain/execution/chaincode"
)

var codeRegistryAddr = bytes.Repeat([]byte{0}, 32)

type CodeDriver interface {
	// Install is called when code deployment transaction is received
	// Example data field - download url for code binary
	// After successful Install, getInstance should give a Chaincode instance without error
	Install(codeID, data []byte) error
	GetInstance(codeID []byte) (chaincode.Chaincode, error)
}

type DriverType uint8

const (
	DriverTypeNative DriverType = iota + 1
)

type CodeInfo struct {
	DriverType DriverType `json:"driverType"`
	CodeID     []byte     `json:"codeID"`
}

type codeRegistry struct {
	drivers map[DriverType]CodeDriver
}

func newCodeRegistry() *codeRegistry {
	reg := new(codeRegistry)
	reg.drivers = make(map[DriverType]CodeDriver)
	return reg
}

func (reg *codeRegistry) registerDriver(driverType DriverType, driver CodeDriver) error {
	if _, found := reg.drivers[driverType]; found {
		return errors.New("driver already registered")
	}
	reg.drivers[driverType] = driver
	return nil
}

func (reg *codeRegistry) install(input *DeploymentInput) error {
	driver, err := reg.getDriver(input.CodeInfo.DriverType)
	if err != nil {
		return err
	}
	return driver.Install(input.CodeInfo.CodeID, input.InstallData)
}

func (reg *codeRegistry) deploy(
	codeAddr []byte, input *DeploymentInput, state State,
) (chaincode.Chaincode, error) {
	driver, err := reg.getDriver(input.CodeInfo.DriverType)
	if err != nil {
		return nil, err
	}
	reg.setCodeInfo(codeAddr, &input.CodeInfo, state)
	return driver.GetInstance(input.CodeInfo.CodeID)
}

func (reg *codeRegistry) getInstance(codeAddr []byte, state StateRO) (chaincode.Chaincode, error) {
	info, err := reg.getCodeInfo(codeAddr, state)
	if err != nil {
		return nil, err
	}
	driver, err := reg.getDriver(info.DriverType)
	if err != nil {
		return nil, err
	}
	return driver.GetInstance(info.CodeID)
}

func (reg *codeRegistry) getDriver(driverType DriverType) (CodeDriver, error) {
	driver, ok := reg.drivers[driverType]
	if !ok {
		return nil, errors.New("unknown chaincode driver type")
	}
	return driver, nil
}

func (reg *codeRegistry) setCodeInfo(codeAddr []byte, codeInfo *CodeInfo, state State) error {
	b, err := json.Marshal(codeInfo)
	if err != nil {
		return err
	}
	state.SetState(codeAddr, b)
	return nil
}

func (reg *codeRegistry) getCodeInfo(codeAddr []byte, state StateRO) (*CodeInfo, error) {
	b := state.GetState(codeAddr)
	info := new(CodeInfo)
	err := json.Unmarshal(b, info)
	if err != nil {
		return nil, err
	}
	return info, nil
}
