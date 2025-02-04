package scheduler

import (
	"github.com/cbthchbc/determisticExecution/application"
	"github.com/cbthchbc/determisticExecution/backend"
	"github.com/cbthchbc/determisticExecution/common"
	"github.com/cbthchbc/determisticExecution/proto_"
	"unsafe"
)

const NumThreads = 4

type socketT struct{}
type messageT struct{}

type DeterministicSchedulerImpl interface {
	DeterministicScheduler(conf *common.Configuration, batchConnection *common.Connection,
		storage *backend.Storage, application application.Application)

	// Function for starting main loops in a separate pthreads.

	RunWorkerThread(arg unsafe.Pointer) unsafe.Pointer

	LockManagerThread(arg unsafe.Pointer) unsafe.Pointer

	SendTxnPtr(socket *socketT, txn *proto_.TxnProto)

	GetTxnPtr(socket *socketT, msg *messageT) *proto_.TxnProto
}

type DeterministicScheduler struct {

	// Configuration specifying node & system settings.
	configuration_ *common.Configuration

	// Thread contexts and their associated Connection objects.
	Threads_          [NumThreads]chan bool
	ThreadConnections [NumThreads]*common.Connection

	lockManagerThread chan bool

	// Connection for receiving txn batches from sequencer.
	BatchConnection *common.Connection

	// Storage layer used in application execution.
	Storage_ *backend.Storage

	// Application currently being run.
	Application_ *application.Application // TODO:const

	// The per-node lock manager tracks what transactions have temporary ownership
	// of what database objects, allowing the scheduler to track LOCAL conflicts
	// and enforce equivalence to transaction orders.
	LockManager *DeterministicLockManager

	// Queue of transaction ids of transactions that have acquired all locks that
	// they have requested.
	//std::deque<TxnProto*>* ready_txns_;
	ReadyTxns *[]*proto_.TxnProto

	// Sockets for communication between main scheduler thread and worker threads.
	//requestsOut  *net.Conn
	//requestsIn   *net.Conn
	//responsesOut [NumThreads]*net.Conn
	//responsesIn  *net.Conn

	// TODO: AtomicQueue
	TxnsQueue     *[]*proto_.TxnProto
	DoneQueue     *[]*proto_.TxnProto
	MessageQueues *[NumThreads]proto_.MessageProto
}

func (d *DeterministicScheduler) Close() error {
	return nil
}
