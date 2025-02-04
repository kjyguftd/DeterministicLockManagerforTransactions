package application

import (
	"github.com/cbthchbc/determisticExecution/backend"
	"github.com/cbthchbc/determisticExecution/common"
	"github.com/cbthchbc/determisticExecution/proto_"
)

type Impl interface {
	// NewTxn Load generation.
	NewTxn(txnId int64, txnType int, args string, config *common.Configuration) *proto_.TxnProto

	// CheckpointID Static method to convert a key into an int for an array
	CheckpointID(key common.Key) int

	// Execute a transaction's application logic given the input 'txn'.
	Execute(txn *proto_.TxnProto, storage *backend.StorageManager) int

	// InitializeStorage Storage initialization method.
	InitializeStorage(storage *backend.Storage, conf *common.Configuration)
}

type Application struct {
	ApplicationImpl Impl
}
