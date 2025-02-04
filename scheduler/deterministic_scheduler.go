package scheduler

import "C"
import (
	"encoding/binary"
	"fmt"
	"github.com/cbthchbc/determisticExecution/application"
	"github.com/cbthchbc/determisticExecution/backend"
	"github.com/cbthchbc/determisticExecution/common"
	"github.com/cbthchbc/determisticExecution/proto_"
	"github.com/golang/protobuf/proto"
	"gopkg.in/zeromq/goczmq.v4"
	"math/rand"
	"runtime"
	"strconv"
	"sync"
	"time"
	"unsafe"
)

type arg struct {
	thread int
	deter  *DeterministicScheduler
}

const ColdCutoff = 990000 // microbenchmark.cc

// DeleteTxnPtr : TODO
func DeleteTxnPtr(data unsafe.Pointer, hint unsafe.Pointer) {
	C.free(data)
}

func (d *DeterministicScheduler) SendTxnPtr(socket *goczmq.Sock, txn *proto_.TxnProto) {
	txnPtr := &txn
	msg := make([]byte, unsafe.Sizeof(*txnPtr))
	binary.LittleEndian.PutUint64(msg, uint64(uintptr(unsafe.Pointer(txnPtr))))
	defer socket.Destroy()

	err := socket.SendFrame(msg, goczmq.FlagNone)
	if err != nil {
		panic(err)
	}
}

func (d *DeterministicScheduler) GetTxnPtr(socket *goczmq.Sock, msg *[]byte) *proto_.TxnProto {
	if _, err := socket.RecvMessage(); err != nil {
		return nil
	}
	txn := (*proto_.TxnProto)(unsafe.Pointer(msg))
	return txn
}

// NewDeterministicScheduler : TODO
func NewDeterministicScheduler(conf *common.Configuration, batchConnection *common.Connection, storage *backend.Storage, application *application.Application) {
	readyTxns_ := new([]*proto_.TxnProto)
	var messageQueues *[NumThreads]proto_.MessageProto

	d := &DeterministicScheduler{
		configuration_:    conf,
		Threads_:          [4]chan bool{},
		ThreadConnections: [4]*common.Connection{},
		lockManagerThread: nil,
		BatchConnection:   batchConnection,
		Storage_:          storage,
		Application_:      application,
		LockManager:       NewDeterministicLockManager(readyTxns_, conf),
		ReadyTxns:         readyTxns_,
		TxnsQueue:         new([]*proto_.TxnProto),
		DoneQueue:         new([]*proto_.TxnProto),
		MessageQueues:     messageQueues,
	}

	for i := 0; i < NumThreads; i++ {
		messageQueues[i] = *new(proto_.MessageProto)
	}

	time.Sleep(time.Duration(1 * float64(time.Second)))

	runtime.GOMAXPROCS(NumThreads)

	var wg sync.WaitGroup
	wg.Add(NumThreads)

	go func() {
		defer wg.Done()
		d.LockManagerThread(arg{})
	}()

	for i := 0; i < NumThreads; i++ {
		go func() {
			defer wg.Done()
			d.RunWorkerThread(arg{})
		}()
	}
	wg.Wait()
}

func UnFetchALL(storage *backend.Storage, txn *proto_.TxnProto) {
	for i := 0; i < len(txn.ReadSet); i++ {
		if read, _ := strconv.Atoi(string(txn.LocateReadSet(byte(i)))); read < ColdCutoff {
			storage.StorageImpl.Unfetch(txn.LocateReadSet(byte(i)))
		}
	}

	for i := 0; i < len(txn.ReadWriteSet); i++ {
		if read, _ := strconv.Atoi(string(txn.LocateReadWriteSet(byte(i)))); read < ColdCutoff {
			storage.StorageImpl.Unfetch(txn.LocateReadWriteSet(byte(i)))
		}
	}

	for i := 0; i < len(txn.WriteSet); i++ {
		if read, _ := strconv.Atoi(string(txn.LocateWriteSet(byte(i)))); read < ColdCutoff {
			storage.StorageImpl.Unfetch(txn.LocateWriteSet(byte(i)))
		}
	}
}

func (d *DeterministicScheduler) RunWorkerThread(arg arg) {
	thread := arg.thread
	scheduler := arg.deter

	activeTxns := make(map[string]*backend.StorageManager)

	// Begin main loop.
	for {
		message := scheduler.MessageQueues[len(scheduler.MessageQueues)-1]
		scheduler.MessageQueues = (*[4]proto_.MessageProto)(scheduler.MessageQueues[:len(scheduler.MessageQueues)-1])

		if message.DataSize != 0 {
			// Remote read result.
			if message.MessageType == 3 { //READ_RESULT
				manager := activeTxns[message.DestinationChannel]
				manager.StorageManagerImpl.HandleReadResult(message)
				if manager.ReadyToExecute() {
					// Execute and clean up.
					txn := manager.Txn_
					scheduler.Application_.ApplicationImpl.Execute(txn, manager)

					scheduler.ThreadConnections[thread].UnlinkChannel(strconv.FormatInt(txn.TxnId, 10))
					delete(activeTxns, message.DestinationChannel)
					// Respond to scheduler;
					//scheduler.SendTxnPtr(scheduler.responses_out_[thread], txn);
					*scheduler.DoneQueue = append(*scheduler.DoneQueue, txn)
				}
			}
		} else {
			// No remote read result found, start on next txn if one is waiting.
			var txn *proto_.TxnProto

			txn = (*scheduler.TxnsQueue)[len(*scheduler.TxnsQueue)-1]
			*scheduler.TxnsQueue = (*scheduler.TxnsQueue)[:len(*scheduler.TxnsQueue)-1]

			if txn.TxnId != 0 {
				// Create manager.
				manager := backend.NewStorageManager(scheduler.configuration_, scheduler.ThreadConnections[thread], scheduler.Storage_, txn)

				// Writes occur at this node.
				if manager.ReadyToExecute() {
					// No remote reads. Execute and clean up.
					scheduler.Application_.ApplicationImpl.Execute(txn, manager)

					// Respond to scheduler;
					//scheduler.SendTxnPtr(scheduler.responses_out_[thread], txn);
					*scheduler.DoneQueue = append(*scheduler.DoneQueue, txn)
				} else {
					scheduler.ThreadConnections[thread].LinkChannel(strconv.FormatInt(txn.TxnId, 10))
					// There are outstanding remote reads.
					activeTxns[strconv.FormatInt(txn.TxnId, 10)] = manager
				}
			}
		}
	}
}

var batches map[int]*proto_.MessageProto

func GetBatch(batchId int, connection *common.Connection) *proto_.MessageProto {
	if batches[batchId] != nil {
		// Requested batch has already been received.
		batch := batches[batchId]
		delete(batches, batchId)
		return batch
	} else {
		message := new(proto_.MessageProto)

		for connection.GetMessage(message) {
			if message.MessageType == 2 { //TXN_BATCH
				if int(message.BatchNumber) == batchId {
					return message
				} else {
					batches[int(message.BatchNumber)] = message
					message = &proto_.MessageProto{}
				}
			}
		}
	}
	return nil
}

func (d *DeterministicScheduler) LockManagerThread(arg arg) {
	scheduler := arg.deter

	var batchMessage *proto_.MessageProto
	txns := 0
	times := time.Now()
	executingTxns := 0
	pendingTxns := 0
	batchOffset := 0
	batchNumber := 0

	for {
		var doneTxn *proto_.TxnProto
		doneTxn = (*scheduler.DoneQueue)[len(*scheduler.DoneQueue)-1]
		if doneTxn.TxnId != 0 {
			// We have received a finished transaction back, release the lock
			scheduler.LockManager.Release(doneTxn)
			executingTxns--

			if len(doneTxn.Writers) == 0 || rand.Int()%len(doneTxn.Writers) == 0 {
				txns++
			}

		} else {
			// Have we run out of txns in our batch? Let's get some new ones.
			if batchMessage == nil {
				batchMessage = GetBatch(batchNumber, scheduler.BatchConnection)

				// Done with current batch, get next.
			} else if batchOffset >= batchMessage.DataSize {
				batchOffset = 0
				batchNumber++

				batchMessage = GetBatch(batchNumber, scheduler.BatchConnection)

				// Current batch has remaining txns, grab up to 10.
			} else if executingTxns+pendingTxns < 2000 {

				for i := 0; i < 100; i++ {
					if batchOffset >= batchMessage.DataSize {
						// Oops we ran out of txns in this batch. Stop adding txns for now.
						break
					}
					txn := new(proto_.TxnProto)
					_ = proto.Unmarshal(batchMessage.Data[batchOffset], txn)

					batchOffset++

					scheduler.LockManager.Lock(txn)
					pendingTxns++
				}

			}
		}

		// Start executing any and all ready transactions to get them off our plate
		for len(*scheduler.ReadyTxns) != 0 {
			txn := (*scheduler.ReadyTxns)[0]
			*scheduler.ReadyTxns = (*scheduler.ReadyTxns)[1:len(*scheduler.ReadyTxns)]

			pendingTxns--
			executingTxns++

			*scheduler.TxnsQueue = append(*scheduler.TxnsQueue, txn)
		}

		if time.Now().After(times.Add(1 * time.Second)) {
			totalTime := time.Now().Sub(times).Seconds()
			fmt.Printf("Completed %.2f txns/sec, %d executing, %d pending\n", float64(txns)/totalTime, executingTxns, pendingTxns)
			// Reset txn count.
			times = time.Now()
			txns = 0
		}
	}
}
