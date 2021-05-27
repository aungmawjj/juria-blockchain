// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package core

import (
	"math"
)

// ValidatorStore godoc
type ValidatorStore interface {
	ValidatorCount() int
	MajorityCount() int
	IsValidator(pubKey *PublicKey) bool
	GetValidator(idx int) *PublicKey
	GetValidatorIndex(pubKey *PublicKey) int
}

type simpleValidatorStore struct {
	validators []*PublicKey
	vMap       map[string]int
}

var _ ValidatorStore = (*simpleValidatorStore)(nil)

func NewValidatorStore(validators []*PublicKey) ValidatorStore {
	store := &simpleValidatorStore{
		validators: validators,
	}
	store.vMap = make(map[string]int, len(store.validators))
	for i, v := range store.validators {
		store.vMap[v.String()] = i
	}
	return store
}

func (store *simpleValidatorStore) ValidatorCount() int {
	return len(store.validators)
}

func (store *simpleValidatorStore) MajorityCount() int {
	return MajorityCount(len(store.validators))
}

func (store *simpleValidatorStore) IsValidator(pubKey *PublicKey) bool {
	_, ok := store.vMap[pubKey.String()]
	return ok
}

func (store *simpleValidatorStore) GetValidator(idx int) *PublicKey {
	if idx >= len(store.validators) || idx < 0 {
		return nil
	}
	return store.validators[idx]
}

func (store *simpleValidatorStore) GetValidatorIndex(pubKey *PublicKey) int {
	return store.vMap[pubKey.String()]
}

// MajorityCount returns 2f + 1 members
func MajorityCount(validatorCount int) int {
	// n=3f+1 -> f=floor((n-1)3) -> m=n-f -> m=ceil((2n+1)/3)
	return int(math.Ceil(float64(2*validatorCount+1) / 3))
}
