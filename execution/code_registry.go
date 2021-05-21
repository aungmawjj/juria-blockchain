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

type codeDeployment struct {
	codeAddr    []byte
	codeInfo    CodeInfo
	installData []byte
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

func (reg *codeRegistry) deploy(dep *codeDeployment, state State) (chaincode.Chaincode, error) {
	driver, err := reg.getDriver(dep.codeInfo.DriverType)
	if err != nil {
		return nil, err
	}
	if err := driver.Install(dep.codeInfo.CodeID, dep.installData); err != nil {
		return nil, err
	}
	reg.setCodeInfo(dep.codeAddr, &dep.codeInfo, state)
	return driver.GetInstance(dep.codeInfo.CodeID)
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
	return info, json.Unmarshal(b, info)
}
