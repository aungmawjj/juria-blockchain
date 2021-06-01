// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package bincc

import (
	"encoding/binary"
	"fmt"
	"io"
)

type readWriter struct {
	reader io.ReadCloser
	writer io.WriteCloser
}

func (rw *readWriter) write(b []byte) error {
	payload := make([]byte, 4, 4+len(b))
	binary.BigEndian.PutUint32(payload, uint32(len(b)))
	payload = append(payload, b...)

	_, err := rw.writer.Write(payload)
	return err
}

func (rw *readWriter) read() ([]byte, error) {
	b, err := rw.readFixedSize(4)
	if err != nil {
		return nil, err
	}
	size := binary.BigEndian.Uint32(b)
	if size > MessageSizeLimit {
		return nil, fmt.Errorf("big message size %d", size)
	}
	return rw.readFixedSize(size)
}

func (rw *readWriter) readFixedSize(size uint32) ([]byte, error) {
	b := make([]byte, size)
	_, err := io.ReadFull(rw.reader, b)
	return b, err
}
