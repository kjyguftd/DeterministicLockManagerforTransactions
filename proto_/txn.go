package proto_

type IsolationLevel int

const (
	SERIALIZABLE IsolationLevel = iota
	SNAPSHOT
	READ_COMMITTED
	READ_UNCOMMITTED
)

type Status int

const (
	NEW Status = iota
	ACTIVE
	COMMITTED
	ABORTED
)

type TxnProto struct {
	// Globally unique transaction id, specifying global order.
	TxnId int64 `protobuf:"variant,1,name=txn_id,json=txnId,proto3" json:"txn_id,omitempty"`

	// Specifies which stored procedure to invoke at execution time.
	TxnType int32 `protobuf:"variant,10,name=txn_type,json=txnType,proto3" json:"txn_type,omitempty"`

	// Isolation level at which to execute transaction.
	// Note: Currently only full serializability is supported.
	IsolationLevel IsolationLevel `protobuf:"variant,11,opt,name=isolation_level,json=isolationLevel,proto3,enum=main.TxnProto_IsolationLevel" json:"isolation_level,omitempty"`

	// True if transaction is known to span multiple database nodes.
	MultiPartition bool `protobuf:"variant,12,opt,name=multipartition,proto3" json:"multipartition,omitempty"`

	// Keys of objects read (but not modified) by this transaction.
	ReadSet [][]byte `protobuf:"bytes,20,rep,name=read_set,json=readSet,proto3" json:"read_set,omitempty"`

	// Keys of objects modified (but not read) by this transaction.
	WriteSet [][]byte `protobuf:"bytes,21,rep,name=write_set,json=writeSet,proto3" json:"write_set,omitempty"`

	// Keys of objects read AND modified by this transaction.
	ReadWriteSet [][]byte `protobuf:"bytes,22,rep,name=read_write_set,json=readWriteSet,proto3" json:"read_write_set,omitempty"`

	// Arguments to be passed when invoking the stored procedure to execute this
	// transaction. 'arg' is a serialized protocol message. The client and backend
	// application code is assumed to know how to interpret this protocol message
	// based on 'txn_type'.
	Arg []byte `protobuf:"bytes,23,opt,name=arg,proto3" json:"arg,omitempty"`

	// Transaction status.
	Status Status `protobuf:"variant,30,opt,name=status,proto3,enum=main.TxnProto_Status" json:"status,omitempty"`

	// Node ids of nodes that participate as readers and writers in this txn.
	Readers []int32 `protobuf:"variant,40,rep,name=readers,proto3" json:"readers,omitempty"`
	Writers []int32 `protobuf:"variant,41,rep,name=writers,proto3" json:"writers,omitempty"`
}

func (txn *TxnProto) Reset() {
	//TODO implement me
	panic("implement me")
}

func (txn *TxnProto) String() string {
	//TODO implement me
	panic("implement me")
}

func (txn *TxnProto) ProtoMessage() {
	//TODO implement me
	panic("implement me")
}

func (txn *TxnProto) LocateReadWriteSet(b byte) []byte {
	return txn.ReadWriteSet[b]
}

func (txn *TxnProto) LocateReadSet(b byte) []byte {
	return txn.ReadSet[b]
}
func (txn *TxnProto) LocateWriteSet(b byte) []byte {
	return txn.WriteSet[b]
}

func (txn *TxnProto) LocateReaders(b byte) int32 {
	return txn.Readers[b]
}
