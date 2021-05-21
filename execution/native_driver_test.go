// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package execution

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNativeCodeDriver(t *testing.T) {
	assert := assert.New(t)

	drv := newNativeCodeDriver()
	cc, err := drv.GetInstance([]byte(NativeCodeIDJuriaCoin))

	assert.NoError(err)
	assert.NotNil(cc)

	err = drv.Install([]byte(NativeCodeIDJuriaCoin), nil)

	assert.NoError(err)
}
