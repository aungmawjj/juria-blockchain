// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package execution

// stateVerifier is used for state query calls
// it calls the VerifyState of state store instead of GetState
// to verify the state value with the merkle root
type stateVerifier struct {
	store     StateStore
	keyPrefix []byte
}

func newStateVerifier(store StateStore, prefix []byte) *stateVerifier {
	return &stateVerifier{
		store:     store,
		keyPrefix: prefix,
	}
}

func (sv *stateVerifier) GetState(key []byte) []byte {
	key = concatBytes(sv.keyPrefix, key)
	return sv.store.VerifyState(key)
}
