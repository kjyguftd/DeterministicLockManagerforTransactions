// Lock manager implementing deterministic two-phase locking as described in
// 'The Case for Determinism in Database Systems'.

package scheduler

import (
	"container/list"
	"github.com/cbthchbc/determisticExecution/common"
	"github.com/cbthchbc/determisticExecution/proto_"
)

const TableSize = 1000000

// The DeterministicLockManager's lock table tracks all lock requests. For a
// given key, if 'lock_table_' contains a nonempty queue, then the item with
// that key is locked and either:
//  (a) first element in the queue specifies the owner if that item is a
//      request for a White lock, or
//  (b) a read lock is held by all elements of the longest prefix of the queue
//      containing only read lock requests.
// Note: using STL deque rather than queue for erase(iterator position).

type LockRequest struct {
	Txn  *proto_.TxnProto // Pointer to txn requesting the lock.
	Mode LockMode         // Specifies whether this is a read or write lock request.
}

type KeysList struct {
	Key          common.Key
	LockRequests []*LockRequest
	//TODO
}

type DeterministicManagerImpl interface {
	DeterministicLockManagerImpl(readyTxns *list.List, config *common.Configuration)
	Lock(txn *proto_.TxnProto) int
	ReleaseByKey(key common.Key, txn *proto_.TxnProto) //const Key& key
	Release(txn *proto_.TxnProto)
}

type DeterministicLockManager struct {
	// Queue of pointers to transactions that have acquired all locks that
	// they have requested. 'ready_txns_[key].front()' is the owner of the lock
	// for a specified key.
	//
	// Owned by the DeterministicScheduler.
	ReadyTxns *[]*proto_.TxnProto
	// TODO: replace array with slice?

	// Configuration object (needed to avoid locking non-local keys).
	Configuration *common.Configuration

	LockTable []*list.List

	// Tracks all txns still waiting on acquiring at least one lock. Entries in
	// 'txn_waits_' are invalided by any call to Release() with the entry's
	// txn.
	TxnWaits map[*proto_.TxnProto]int
}

func NewDeterministicLockManager(readyTxn *[]*proto_.TxnProto, config *common.Configuration) *DeterministicLockManager {
	lt := newLockTable()
	return &DeterministicLockManager{
		ReadyTxns:     readyTxn,
		Configuration: config,
		LockTable:     lt,
		TxnWaits:      make(map[*proto_.TxnProto]int),
		//DeterManagerImpl: DeterministicManagerImpl
	}
}

func newLockTable() []*list.List {
	lt := make([]*list.List, TableSize)
	for i := 0; i < TableSize; i++ {
		e := list.New()
		e.PushBack(&KeysList{})
		lt[i] = e
	}
	return lt
}

func NewLockRequest(txn *proto_.TxnProto, Mode LockMode) *LockRequest {
	return &LockRequest{
		Txn:  txn,
		Mode: Mode,
	}
}

func NewKeysList(key common.Key, lockRequest []*LockRequest) *KeysList {
	return &KeysList{
		Key:          key,
		LockRequests: lockRequest,
	}
}

func (d *DeterministicLockManager) Close() error {
	return nil
}

func (d *DeterministicLockManager) hash(key common.Key) int {
	hash := uint64(2166136261)
	for i := 0; i < len(key); i++ {
		hash = hash ^ uint64(key[i])
		hash = hash * 16777619
	}
	return int(hash % TableSize)
}

func (d *DeterministicLockManager) isLocal(key common.Key) bool {
	return d.Configuration.ConfigImpl.LookupPartition(key) == d.Configuration.NodeID
}
