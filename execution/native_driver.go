// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package execution

import (
	"bytes"
	"errors"

	"github.com/aungmawjj/juria-blockchain/execution/chaincode"
	"github.com/aungmawjj/juria-blockchain/execution/chaincode/juriacoin"
)

var (
	NativeCodeIDJuriaCoin = string(bytes.Repeat([]byte{1}, 32))
)

type nativeCodeDriver struct{}

var _ CodeDriver = (*nativeCodeDriver)(nil)

func newNativeCodeDriver() *nativeCodeDriver {
	return new(nativeCodeDriver)
}

func (drv *nativeCodeDriver) Install(codeID, data []byte) error {
	_, err := drv.GetInstance(codeID)
	return err
}

func (drv *nativeCodeDriver) GetInstance(codeID []byte) (chaincode.Chaincode, error) {
	switch string(codeID) {
	case NativeCodeIDJuriaCoin:
		return new(juriacoin.JuriaCoin), nil
	default:
		return nil, errors.New("unknown native chaincode id")
	}
}
