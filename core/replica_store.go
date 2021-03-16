// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package core

// ReplicaStore godoc
type ReplicaStore interface {
	ReplicaCount() int
	IsReplica(pubKey *PublicKey) bool
	GetReplica(idx int) []byte
	GetReplicaIndex(pubKey *PublicKey) (int, bool)
}
