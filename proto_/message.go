package proto_

type MessageType int

// TODO: repeated & optional

const (
	EMPTY MessageType = iota
	TXN_PROTO
	TXN_BATCH
	READ_RESULT
	LINK_CHANNEL   // [Connection implementation specific.]
	UNLINK_CHANNEL // [Connection implementation specific.]
	TXN_PTR
	MESSAGE_PTR
)

type MessageProto struct {
	// Node to which this message should be sent.
	DestinationNode int32 `protobuf:"variant,1,opt,name=destination_txn,json=destinationTxn,proto3" json:"destination_node,omitempty"`

	// Channel to which this message shall be delivered when it arrives at node
	// 'destination_node'.
	DestinationChannel string `protobuf:"variant, 2, opt, name= destinationChannel, json=destinationChannel,proto3" json:"destination_channel,omitempty"`

	// Node from which the message originated.
	SourceNode int32 `protobuf:"variant, 3, opt, name= sourceNode, json=sourceNode,proto3" json:"source_node,omitempty"`

	// Channel from which the message originated
	SourceChannel int32 `protobuf:"variant, 4, opt, name=sourceChannel, json=sourceChannel,proto3" json:"source_channel,omitempty"'`

	// Every type of network message should get an entry here
	MessageType MessageType `protobuf:"variant, 9, opt, name=messageType, json=messageType,proto3" json:"message_type,omitempty"`

	// Actual data for the message being carried, to be deserialized into a
	// protocol message object of type depending on 'type'. In TXN_PROTO and
	// TXN_BATCH messages, 'data' contains are one and any number of TxnProtos,
	// respectively.
	Data [][]byte `protobuf:"variant, 11, opt, name=data, json=data, proto3" json:"data,omitempty"`

	DataSize int `protobuf:"variant, 11, opt, name=datasize, json=datasize, proto3" json:"datasize,omitempty"`

	// Pointer to actual data for message being carried. Can only be used for
	// messages between threads.
	DataPtr int64 `protobuf:"variant, 12, opt, name=dataPtr, json=dataPtr, proto3" json:"data_ptr,omitempty"`

	// For TXN_BATCH messages, 'batch_number' identifies the epoch of the txn
	// batch being sent
	BatchNumber int64 `protobuf:"variant, 21, opt, name=batchNumber, json=batchNumber,proto3" json:"batch_number,omitempty"`

	// For READ_RESULT messages, 'keys(i)' and 'values(i)' store the key and
	// result of a read, respectively.
	Keys   [][]byte `protobuf:"byte, 31, rep, name=keys, json=keys,proto3" json:"keys,omitempty"`
	Values [][]byte `protobuf:"byte, 32, rep, name=values, json=values,proto3" json:"values,omitempty"`

	// For (UN)LINK_CHANNEL messages, specifies the main channel of the requesting
	// Connection object.
	MainChannel string `protobuf:"variant, 1001, opt, name=mainChannel, json=mainChannel,proto3" json:"main_channel,omitempty"`

	// For (UN)LINK_CHANNEL messages, specifies the channel to be (un)linked
	// to the requesting Connection object.
	ChannelRequest string `protobuf:"variant, 1002, opt, name=channelRequest, json=channelRequest,proto3" json:"channel_request,omitempty"`
}
