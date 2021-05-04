// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package util

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
)

var ByteOrder binary.ByteOrder = binary.BigEndian

func ConcatBytes(srcs ...[]byte) []byte {
	buf := bytes.NewBuffer(nil)
	for _, src := range srcs {
		buf.Grow(len(src))
	}
	for _, src := range srcs {
		buf.Write(src)
	}
	return buf.Bytes()
}

func Base64String(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}
