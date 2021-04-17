// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package core

type Marshaler interface {
	Marshal() ([]byte, error)
}

type Unmarshaler interface {
	Unmarshal(b []byte) error
}
