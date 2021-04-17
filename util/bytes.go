// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package util

func ConcatBytes(srcs ...[]byte) []byte {
	size := 0
	for _, src := range srcs {
		size += len(src)
	}
	dst := make([]byte, 0, size)
	for _, src := range srcs {
		dst = append(dst, src...)
	}
	return dst
}
