package backend

import (
	"flag"
	"github.com/cbthchbc/determisticExecution/common"
	"github.com/cbthchbc/determisticExecution/proto_"
	"reflect"
	"strconv"
	"unsafe"
)

type StorageManagerImpl interface {

	// TODO(alex): Document this class correctly.

	ReadObject(key common.Key) *flag.Value
	PutObject(key common.Key, value *flag.Value) bool
	DeleteObject(key common.Key) bool
	HandleReadResult(message proto_.MessageProto)
	ReadyToExecute() bool
	GetStorage() *Storage
	//{
	//return actual_storage_
	//}
}

type StorageManager struct {

	// Set by the constructor, indicating whether 'txn' involves any writes at
	// this node.
	Writer bool

	// private:
	//friend class DeterministicScheduler;

	// Pointer to the configuration object for this node.
	configuration_ *common.Configuration

	// A Connection object that can be used to send and receive messages.
	connection_ *common.Connection

	// Storage layer that *actually* stores data objects on this node.
	actualStorage_ *Storage

	// Transaction that corresponds to this instance of a StorageManager.
	Txn_ *proto_.TxnProto

	// Local copy of all data objects read/written by 'txn_', populated at
	// StorageManager construction time.
	//
	// TODO(alex): Should these be pointers to reduce object copying overhead?
	objects_ map[*byte]*byte

	remoteReads []*flag.Value //TODO:VALUE???

	StorageManagerImpl StorageManagerImpl
}

func NewStorageManager(configuration_ *common.Configuration, connection_ *common.Connection, actualStorage_ *Storage, txn_ *proto_.TxnProto) *StorageManager {
	s := &StorageManager{
		configuration_: configuration_,
		connection_:    connection_,
		actualStorage_: actualStorage_,
		Txn_:           txn_,
		objects_:       make(map[*byte]*byte),
	}

	messageProto := proto_.MessageProto{}

	reader := false
	for i := 0; i < len(txn_.Readers); i++ {
		if int(txn_.LocateReaders(byte(i))) == configuration_.NodeID {
			reader = true
		}
	}

	if reader {
		messageProto.DestinationChannel = strconv.Itoa(int(txn_.TxnId))
		messageProto.MessageType = proto_.MessageType(3) //READ_RESULT

		// Execute local reads
		for i := 0; i < len(txn_.ReadSet); i++ {
			key := txn_.LocateReadSet(byte(i))
			if configuration_.ConfigImpl.LookupPartition(key) == configuration_.NodeID {
				val := actualStorage_.StorageImpl.ReadObject(key, 0)
				s.objects_[&key[0]] = val
				messageProto.Keys = append(messageProto.Keys, key)
				if val == nil {
					messageProto.Values = append(messageProto.Values, nil)
				} else {
					messageProto.Values = append(messageProto.Values, BytePtrToByteSlice(val))
				}
			}
		}
		for i := 0; i < len(txn_.ReadWriteSet); i++ {
			key := txn_.LocateReadWriteSet(byte(i))
			if configuration_.ConfigImpl.LookupPartition(key) == configuration_.NodeID {
				val := actualStorage_.StorageImpl.ReadObject(key, 0)
				s.objects_[&key[0]] = val
				messageProto.Keys = append(messageProto.Keys, key)
				if val == nil {
					messageProto.Values = append(messageProto.Values, nil)
				} else {
					messageProto.Values = append(messageProto.Values, BytePtrToByteSlice(val))
				}
			}
		}
		for i := 0; i < len(txn_.Writers); i++ {
			if int(txn_.Writers[i]) != configuration_.NodeID {
				messageProto.DestinationNode = txn_.Writers[i]
				connection_.Send1(messageProto) // TODO
			}
		}
	}
	s.Writer = false
	for i := 0; i < len(txn_.Writers); i++ {
		if int(txn_.Writers[i]) == configuration_.NodeID {
			s.Writer = true
		}
	}

	return s
}

func BytePtrToByteSlice(bytePtr *byte) []byte {
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&bytePtr))
	return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: sliceHeader.Data,
		Len:  sliceHeader.Len,
		Cap:  sliceHeader.Len,
	}))
}

//func (s *StorageManager) HandleReadResult(message *MessageProto) {
//	assert(message.GetType() == MessageProto_READ_RESULT)
//	for i := 0; i < len(message.GetKeys()); i++ {
//		val := NewValue(message.GetValues()[i])
//		s.objects_[message.GetKeys()[i]] = val
//		s.remoteReads = append(s.remoteReads, val)
//	}
//}

func (s *StorageManager) ReadyToExecute() bool {
	return len(s.objects_) == len(s.Txn_.ReadSet)+len(s.Txn_.ReadWriteSet) // TODO:size &len
}

//func (s *StorageManager) ReadObject(key common.Key) *byte {
//	//return s.objects_[key]
//}

func (s *StorageManager) PutObject(key common.Key, value *byte) bool {
	// Write object to storage if applicable.
	if s.configuration_.ConfigImpl.LookupPartition(key) == s.configuration_.NodeID {
		return s.actualStorage_.StorageImpl.PutObject(key, value, s.Txn_.TxnId)
	} else {
		return true // Not this node's problem.
	}
}

func (s *StorageManager) DeleteObject(key common.Key) bool {
	// Delete object from storage if applicable.
	if s.configuration_.ConfigImpl.LookupPartition(key) == s.configuration_.NodeID {
		return s.actualStorage_.StorageImpl.DeleteObject(key, s.Txn_.TxnId) //TODO:gettxnid
	} else {
		return true // Not this node's problem.
	}
}
