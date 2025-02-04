package scheduler

import (
	"fmt"
	"github.com/cbthchbc/determisticExecution/application"
	"github.com/cbthchbc/determisticExecution/backend"
	"github.com/cbthchbc/determisticExecution/common"
	"github.com/cbthchbc/determisticExecution/proto_"
	"github.com/golang/protobuf/proto"
	"os"
	"strconv"
	"time"
)

func (s *SerialScheduler) NewSerialScheduler(conf *common.Configuration, connection *common.Connection, storage *backend.Storage, checkpointing bool) *SerialScheduler {
	return &SerialScheduler{
		configuration: conf,
		connection:    connection,
		storage:       storage,
		checkpointing: checkpointing,
	}
}

func (s *SerialScheduler) Run(application *application.Application) {
	var message proto_.MessageProto
	var txn proto_.TxnProto
	managerConnection := s.connection.Multiplexer.NewConnection("manager_connection")
	defer managerConnection.Close()

	txns := 0
	times := time.Now()
	startTime := times
	for {
		if s.connection.GetMessage(&message) {
			// Execute all txns in batch.
			for i := 0; i < message.DataSize; i++ {
				_ = proto.Unmarshal(message.Data[i], &txn)

				// Link txn-specific channel ot manager_connection.
				managerConnection.LinkChannel(strconv.FormatInt(txn.TxnId, 10))

				// Create manager.
				manager := backend.NewStorageManager(s.configuration, managerConnection, s.storage, &txn)

				// Execute txn if any writes occur at this node.
				if manager.Writer == true {
					for !manager.ReadyToExecute() {
						if s.connection.GetMessage(&message) {
							manager.StorageManagerImpl.HandleReadResult(message)
						}
					}
					application.ApplicationImpl.Execute(&txn, manager)
				}

				// Clean up the mess.
				managerConnection.UnlinkChannel(strconv.FormatInt(txn.TxnId, 10))

				// Report throughput (once per second). TODO(alex): Fix reporting.
				if txn.Writers[txn.TxnId%int64(len(txn.Writers))] == int32(s.configuration.NodeID) {
					txns++
				}

				if time.Now().After(times.Add(1 * time.Second)) {
					fmt.Printf("Executed %d txns\n", txns)
					// Reset txn count.
					times = time.Now()
					txns = 0
				}
			}
		}

		//Report throughput (once per second).
		if time.Now().After(times.Add(1 * time.Second)) {
			fmt.Printf("Executed %d txns\n", txns)
			// Reset txn count.
			times = time.Now()
			txns = 0
		}

		// Run for at most one minute.
		if time.Now().After(startTime.Add(60 * time.Second)) {
			os.Exit(0)
		}
	}
}
