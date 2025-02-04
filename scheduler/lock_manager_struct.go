package scheduler

import (
	"github.com/cbthchbc/determisticExecution/common"
	"github.com/cbthchbc/determisticExecution/proto_"
)

type LockMode int

const (
	UNLOCKED LockMode = iota
	READ
	WRITE
)

type LockManager interface {

	// Lock Attempts to assign the lock for each key in keys to the specified
	// transaction. Returns the number of requested locks NOT assigned to the
	// transaction (therefore Lock() returns 0 if the transaction successfully
	// acquires all locks).
	//
	// Requires: 'read_keys' and 'write_keys' do not overlap, and neither contains
	//           duplicate keys.
	// Requires: Lock has not previously been called with this txn_id. Note that
	//           this means Lock can only ever be called once per txn.
	Lock(txn *proto_.TxnProto) int
	// return 0

	// For each key in 'keys':
	//   - If the specified transaction owns the lock on the item, the lock is
	//     released.
	//   - If the transaction is in the queue to acquire a lock on the item, the
	//     request is cancelled and the transaction is removed from the item's
	//     queue.

	ReleaseByKey(key common.Key, txn *proto_.TxnProto)
	Release(txn *proto_.TxnProto)

	// Status Locked sets '*owner' to contain the txn IDs of all txns holding the lock,
	// and returns the current state of the lock: UNLOCKED if it is not currently
	// held, READ or WRITE if it is, depending on the current state.
	Status(key common.Key, owners *[]*proto_.TxnProto) LockMode
}

func (d *LockRequest) Close() error {
	return nil
}
