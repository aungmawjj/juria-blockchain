package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConcactBytes(t *testing.T) {
	assert := assert.New(t)
	res := ConcatBytes([]byte{1, 2, 3}, []byte{4, 5, 6}, []byte{7, 8, 9})
	assert.Equal([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9}, res)
}
