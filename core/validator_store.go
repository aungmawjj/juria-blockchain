// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package core

import "math"

// ValidatorStore godoc
type ValidatorStore interface {
	ValidatorCount() int
	MajorityCount() int
	IsValidator(pubKey *PublicKey) bool
	GetValidator(idx int) []byte
	GetValidatorIndex(pubKey *PublicKey) (int, bool)
}

// majorityCount returns 2f + 1 members
func majorityCount(validatorCount int) int {
	// n=3f+1 -> f=floor((n-1)3) -> m=n-f -> m=ceil((2n+1)/3)
	return int(math.Ceil(float64(2*validatorCount+1) / 3))
}
