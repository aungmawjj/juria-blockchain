// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package execution

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCodeRegistry(t *testing.T) {
	assert := assert.New(t)

	trk := newStateTracker(newMapStateStore(), codeRegistryAddr)
	reg := newCodeRegistry()

	codeAddr := bytes.Repeat([]byte{1}, 32)
	dep := &DeploymentInput{
		CodeInfo: CodeInfo{
			DriverType: DriverTypeNative,
			CodeID:     []byte(NativeCodeIDJuriaCoin),
		},
	}

	cc, err := reg.getInstance(codeAddr, trk)

	assert.Error(err, "code not deployed yet")
	assert.Nil(cc)

	cc, err = reg.deploy(codeAddr, dep, trk)

	assert.Error(err, "native driver not registered yet")
	assert.Nil(cc)

	reg.registerDriver(DriverTypeNative, newNativeCodeDriver())
	cc, err = reg.deploy(codeAddr, dep, trk)

	assert.NoError(err)
	assert.NotNil(cc)

	err = reg.registerDriver(DriverTypeNative, newNativeCodeDriver())

	assert.Error(err, "registered driver twice")

	cc, err = reg.getInstance(codeAddr, trk)

	assert.NoError(err)
	assert.NotNil(cc)

	cc, err = reg.getInstance(bytes.Repeat([]byte{2}, 32), trk)

	assert.Error(err, "wrong code address")
	assert.Nil(cc)
}
