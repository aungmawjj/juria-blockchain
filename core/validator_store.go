// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package core

// ValidatorStore godoc
type ValidatorStore interface {
	ValidatorCount() int
	IsValidator(pubKey *PublicKey) bool
	GetValidator(idx int) []byte
	GetValidatorIndex(pubKey *PublicKey) (int, bool)
}
