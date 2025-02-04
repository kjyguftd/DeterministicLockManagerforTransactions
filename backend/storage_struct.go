package backend

import (
	"github.com/cbthchbc/determisticExecution/common"
	"github.com/cbthchbc/determisticExecution/proto_"
)

type StorageImpl interface {
	Prefetch(key common.Key, waitTime *float64) bool

	Unfetch(key common.Key) bool

	ReadObject(key common.Key, txnID int64) *byte

	PutObject(key common.Key, value *byte, txnID int64) bool

	DeleteObject(key common.Key, txnID int64) bool

	PrepareForCheckpoint(stable int64)
	Checkpoint() int

	InitMutex()
}

type Storage struct {
	Config        *common.Configuration
	Connection    *common.Connection
	ActualStorage *Storage
	Txn           *proto_.TxnProto
	StorageImpl   StorageImpl
}
