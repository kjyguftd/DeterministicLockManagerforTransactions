package common

import (
	"context"
	"github.com/cbthchbc/determisticExecution/proto_"
	"gopkg.in/zeromq/goczmq.v4"
	"net/mail"
	"sync"
)

type ConnectionMultiplexerImpl interface {
	NewConnection(channel string) *Connection

	//NewConnection(channel string, AtomicQueue<MessageProto>** aa) Connection*

}

type ConnectionMultiplexer struct {
	configuration *Configuration
	context       *context.Context
	port          int
	remoteIn      *goczmq.Sock
	remoteOut     map[int]*goczmq.Sock
	inprocIn      *goczmq.Sock
	inprocOut     map[string]*goczmq.Sock
	//remoteResult       map[string]*AtomicQueue
	//linkUnlinkQueue    map[string]*AtomicQueue
	undeliveredMessages       map[string][]proto_.MessageProto
	newConnectionMutex        *sync.Mutex
	sendMutex                 *sync.Mutex
	newConnectionChannel      *string
	deleteConnectionChannel   *string
	newConnection             *Connection
	deconstructorInvoked      bool
	ConnectionMultiplexerImpl ConnectionMultiplexerImpl
}

type Connection struct {
	Channel        string
	LinkedChannels map[string]bool
	Multiplexer    *ConnectionMultiplexer
	SocketOut      *goczmq.Sock
	SocketIn       *goczmq.Sock
	Msg            mail.Message
}

func NewConnectionMultiplexer(config *Configuration) *ConnectionMultiplexer {
	return &ConnectionMultiplexer{
		configuration:           config,
		context:                 nil,
		port:                    0,
		remoteIn:                nil,
		remoteOut:               nil,
		inprocIn:                nil,
		inprocOut:               nil,
		undeliveredMessages:     nil,
		newConnectionMutex:      nil,
		sendMutex:               nil,
		newConnectionChannel:    nil,
		deleteConnectionChannel: nil,
		newConnection:           nil,
		deconstructorInvoked:    false,
	}
}

func (mux *ConnectionMultiplexer) Run() {}

func (mux *ConnectionMultiplexer) Send(message proto_.MessageProto) {}

func NewConnection(channel string) *Connection {
	return &Connection{
		Channel:        channel,
		LinkedChannels: nil,
		Multiplexer:    nil,
		SocketOut:      nil,
		SocketIn:       nil,
		Msg:            mail.Message{},
	}
} //aa **AtomicQueue

func (conn *Connection) Send(message proto_.MessageProto) {}

func (conn *Connection) GetMessage(message *proto_.MessageProto) bool {
	return false
}

func (conn *Connection) GetMessageBlocking(message *proto_.MessageProto, maxWaitTime float64) bool {
	return false
}

func (conn *Connection) LinkChannel(channel string) {}

func (conn *Connection) UnlinkChannel(channel string) {}

func (conn *Connection) Send1(messageProto proto_.MessageProto) {}

func (conn *Connection) Close() {

}
