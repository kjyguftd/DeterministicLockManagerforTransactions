// Lock manager implementing deterministic two-phase locking as described in
// 'The Case for Determinism in Database Systems'.

package scheduler

import (
	"bytes"
	"container/list"
	"github.com/cbthchbc/determisticExecution/common"
	"github.com/cbthchbc/determisticExecution/proto_"
)

func (d *DeterministicLockManager) DeterministicLockManagerImpl(readyTxns *list.List, config *common.Configuration) {
	// TODO implement me
	panic("implement me")
}

func (d *DeterministicLockManager) Lock(txn *proto_.TxnProto) int {
	notAcquired := 0

	// Handle read/write lock requests.
	for i := 0; i < len(txn.ReadWriteSet); i++ {
		// Only lock local keys.
		if d.isLocal(txn.LocateReadWriteSet(byte(i))) {
			keyRequests := d.LockTable[d.hash(txn.LocateReadWriteSet(byte(i)))]

			var keyRequest *list.Element
			for keyRequest = keyRequests.Front(); keyRequest != nil; keyRequest = keyRequest.Next() {
				keysList := keyRequest.Value.(*KeysList)
				if bytes.Equal(keysList.Key, txn.LocateReadWriteSet(byte(i))) {
					break
				}
			}

			var requests []*LockRequest

			if keyRequest == keyRequests.Back() {
				requests = []*LockRequest{}
				keyRequests.PushBack(NewKeysList(txn.LocateReadWriteSet(byte(i)), requests))
			} else {
				requests = keyRequest.Value.(*KeysList).LockRequests
			}

			// Only need to request this if lock txn hasn't already requested it.
			if len(requests) == 0 || txn != requests[len(requests)-1].Txn {
				requests = append(requests, &LockRequest{
					Txn:  txn,
					Mode: WRITE,
				})

				// Write lock request fails if there is any previous request at all.
				if len(requests) > 1 {
					notAcquired++
				}
			}
		}
	}

	// Handle read lock requests. This is last so that we don't have to deal with
	// upgrading lock requests from read to write on hash collisions.
	for i := 0; i < len(txn.ReadSet); i++ {
		// Only lock local keys.
		if d.isLocal(txn.LocateReadSet(byte(i))) {
			keyRequests := d.LockTable[d.hash(txn.LocateReadSet(byte(i)))]

			var keyRequest *list.Element
			for keyRequest = keyRequests.Front(); keyRequest != nil; keyRequest = keyRequest.Next() {
				keysList := keyRequest.Value.(*KeysList)
				if bytes.Equal(keysList.Key, txn.LocateReadSet(byte(i))) {
					break
				}
			}

			var requests []*LockRequest

			if keyRequest == keyRequests.Back() {
				requests = []*LockRequest{}
				keyRequests.PushBack(NewKeysList(txn.LocateReadSet(byte(i)), requests))
			} else {
				requests = keyRequest.Value.(*KeysList).LockRequests
			}

			// Only need to request this if lock txn hasn't already requested it.
			if len(requests) == 0 || txn != requests[len(requests)-1].Txn {
				requests = append(requests, &LockRequest{
					Txn:  txn,
					Mode: READ,
				})
				// Read lock request fails if there is any previous write request.
				for _, request := range requests {
					if request.Mode == WRITE {
						notAcquired++
						break
					}
				}
			}
		}
	}

	// Record and return the number of locks that the txn is blocked on.
	if notAcquired > 0 {
		d.TxnWaits[txn] = notAcquired
	} else {
		*d.ReadyTxns = append(*d.ReadyTxns, txn)
	}
	return notAcquired
}

func (d *DeterministicLockManager) Release(txn *proto_.TxnProto) {
	for i := 0; i < len(txn.ReadSet); i++ {
		if d.isLocal(txn.LocateReadSet(byte(i))) {
			d.ReleaseByKey(txn.LocateReadSet(byte(i)), txn)
		}
	}
	// Currently commented out because nothing in any write set can conflict
	// in TPCC or Microbenchmark.
	//  for (int i = 0; i < txn->write_set_size(); i++)
	//    if (IsLocal(txn->write_set(i)))
	//      Release(txn->write_set(i), txn);
	for i := 0; i < len(txn.ReadWriteSet); i++ {
		if d.isLocal(txn.LocateReadWriteSet(byte(i))) {
			d.ReleaseByKey(txn.LocateReadSet(byte(i)), txn)
		}
	}
}

func (d *DeterministicLockManager) ReleaseByKey(key common.Key, txn *proto_.TxnProto) {
	// Avoid repeatedly looking up key in the unordered_map.
	keyRequests := d.LockTable[d.hash(key)]

	var keyRequest *list.Element
	for keyRequest = keyRequests.Front(); keyRequest != nil && !bytes.Equal(keyRequest.Value.(*KeysList).Key, key); keyRequest = keyRequest.Next() {
	}
	lockRequests := keyRequest.Value.(*KeysList).LockRequests

	// Seek to the target request. Note whether any write lock requests precede
	// the target.
	var writeRequestsPrecedeTarget = false
	var lockRequest *LockRequest
	var i int
	for i, lockRequest = range lockRequests {
		if !(lockRequest.Txn != txn && lockRequest != lockRequests[len(lockRequests)-1]) {
			break
		} else if lockRequest.Mode == WRITE {
			writeRequestsPrecedeTarget = true
		}
	}

	var target *LockRequest
	var q int
	var targetIndex int

	// If we found the request, erase it. No need to do anything otherwise.
	if lockRequest != lockRequests[len(lockRequests)-1] {
		for q, target = range lockRequests[i:] {

			// If there are more requests following the target request, one or more
			// may need to be granted as a result of the target's release.
			if target != lockRequests[len(lockRequests)-1] {
				var newOwners []*proto_.TxnProto

				// Grant subsequent request(s) if:
				//  (a) The canceled request held a write lock.
				//  (b) The canceled request held a read lock ALONE.
				//  (c) The canceled request was a write request preceded only by read
				//      requests and followed by one or more read requests.
				if target == lockRequests[0] && (target.Mode == WRITE || (target.Mode == READ && lockRequest.Mode == WRITE)) {
					// (a) or (b)
					// If a write lock request follows, grant it.
					if lockRequest.Mode == WRITE {
						newOwners = append(newOwners, lockRequest.Txn)
					}

					// If a sequence of read lock requests follows, grant all of them.
					for ; lockRequest != lockRequests[len(lockRequests)-1] && lockRequest.Mode == READ; lockRequest = lockRequests[q+1] {
						newOwners = append(newOwners, lockRequest.Txn)
					}
					targetIndex = q

				} else if !writeRequestsPrecedeTarget && target.Mode == WRITE && lockRequest.Mode == READ {
					// (c)
					// If a sequence of read lock requests follows, grant all of them.
					for ; lockRequest != lockRequests[len(lockRequests)-1] && lockRequest.Mode == READ; lockRequest = lockRequests[q+1] {
						newOwners = append(newOwners, lockRequest.Txn)
					}
					targetIndex = q
				}

				// Handle txns with newly granted requests that may now be ready to run.
				var j uint64 = 0
				for ; j < uint64(len(newOwners)); j++ {
					d.TxnWaits[newOwners[j]]--
					if d.TxnWaits[newOwners[j]] == 0 {
						// The txn that just acquired the released lock is no longer waiting
						// on any lock requests.
						*d.ReadyTxns = append(*d.ReadyTxns, newOwners[j])
						delete(d.TxnWaits, newOwners[j])
					}
				}
			}
		}

		// Now it is safe to actually erase the target request.
		lockRequests = append(lockRequests[:targetIndex], lockRequests[(targetIndex+1):]...)

		if len(lockRequests) == 0 {
			keyRequests.Remove(keyRequest)
		}
	}
}
