package storage

import (
	"github.com/aungmawjj/juria-blockchain/core"
	"github.com/aungmawjj/juria-blockchain/merkle"
	"github.com/dgraph-io/badger/v3"
)

type CommitData struct {
	Block        *core.Block
	Transactions []*core.Transaction
	TxCommits    []*core.TxCommit
	StateChanges []*core.StateChange

	blockCommit  *core.BlockCommit
	merkleUpdate *merkle.UpdateResult
}

type Storage struct {
	db          *badger.DB
	chainStore  *chainStore
	stateStore  *stateStore
	merkleStore *merkleStore
	merkleTree  *merkle.Tree
}

func NewStorage(db *badger.DB, treeOpts merkle.TreeOptions) *Storage {
	strg := new(Storage)
	strg.db = db
	getter := &badgerGetter{db}
	strg.chainStore = &chainStore{getter}
	strg.stateStore = &stateStore{getter, treeOpts.HashFunc}
	strg.merkleStore = &merkleStore{getter}
	strg.merkleTree = merkle.NewTree(strg.merkleStore, treeOpts)
	return strg
}

func (strg *Storage) Commit(data *CommitData) error {
	return strg.commit(data)
}

func (strg *Storage) GetBlock(hash []byte) (*core.Block, error) {
	return strg.chainStore.getBlock(hash)
}

func (strg *Storage) GetLastBlock() (*core.Block, error) {
	return strg.chainStore.getLastBlock()
}

func (strg *Storage) GetBlockHeight() (uint64, error) {
	return strg.chainStore.getBlockHeight()
}

func (strg *Storage) GetBlockByHeight(height uint64) (*core.Block, error) {
	return strg.chainStore.getBlockByHeight(height)
}

func (strg *Storage) GetBlockCommit(hash []byte) (*core.BlockCommit, error) {
	return strg.chainStore.getBlockCommit(hash)
}

func (strg *Storage) GetTx(hash []byte) (*core.Transaction, error) {
	return strg.chainStore.getTx(hash)
}

func (strg *Storage) HasTx(hash []byte) bool {
	return strg.chainStore.hasTx(hash)
}

func (strg *Storage) GetTxCommit(hash []byte) (*core.TxCommit, error) {
	return strg.chainStore.getTxCommit(hash)
}

func (strg *Storage) GetState(key []byte) []byte {
	return strg.stateStore.getState(key)
}

func (strg *Storage) GetMerkleRoot() []byte {
	root := strg.merkleTree.Root()
	if root == nil {
		return nil
	}
	return root.Data
}

func (strg *Storage) commit(data *CommitData) error {
	strg.setMerkleUpdate(data)
	strg.setBlockCommit(data)
	return strg.storeCommitData(data)
}

func (strg *Storage) storeCommitData(data *CommitData) error {
	if err := strg.storeChainData(data); err != nil {
		return err
	}
	if err := strg.storeBlockCommit(data); err != nil {
		return err
	}
	if err := strg.commitStateMerkleTree(data); err != nil {
		return err
	}
	return strg.setCommitedBlockHeight(data.Block.Height())
}

func (strg *Storage) setMerkleUpdate(data *CommitData) {
	strg.stateStore.loadPrevValues(data.StateChanges)
	strg.stateStore.loadPrevTreeIndexes(data.StateChanges)
	prevLeafCount := strg.merkleStore.getLeafCount()
	leafCount := strg.stateStore.setNewTreeIndexes(data.StateChanges, prevLeafCount)
	nodes := strg.stateStore.computeUpdatedTreeNodes(data.StateChanges)
	data.merkleUpdate = strg.merkleTree.Update(nodes, leafCount)
}

func (strg *Storage) setBlockCommit(data *CommitData) {
	data.blockCommit = core.NewBlockCommit().
		SetHash(data.Block.Hash()).
		SetLeafCount(data.merkleUpdate.LeafCount.Bytes()).
		SetStateChanges(data.StateChanges).
		SetMerkleRoot(data.merkleUpdate.Root.Data)
}

func (strg *Storage) storeChainData(data *CommitData) error {
	updFns := make([]updateFunc, 0)
	updFns = append(updFns, strg.chainStore.setBlock(data.Block)...)
	updFns = append(updFns, strg.chainStore.setTxs(data.Transactions)...)
	updFns = append(updFns, strg.chainStore.setTxCommits(data.TxCommits)...)
	return updateBadgerDB(strg.db, updFns)
}

func (strg *Storage) storeBlockCommit(data *CommitData) error {
	updFn := strg.chainStore.setBlockCommit(data.blockCommit)
	return updateBadgerDB(strg.db, []updateFunc{updFn})
}

// commit state values and merkle tree in one transaction
func (strg *Storage) commitStateMerkleTree(data *CommitData) error {
	updFns := strg.stateStore.commitStateChanges(data.StateChanges)
	updFns = append(updFns, strg.merkleStore.commitUpdate(data.merkleUpdate)...)
	return updateBadgerDB(strg.db, updFns)
}

func (strg *Storage) setCommitedBlockHeight(height uint64) error {
	updFn := strg.chainStore.setBlockHeight(height)
	return updateBadgerDB(strg.db, []updateFunc{updFn})
}
