package scheduler

import (
	"github.com/cbthchbc/determisticExecution/application"
	"github.com/cbthchbc/determisticExecution/backend"
	"github.com/cbthchbc/determisticExecution/common"
)

type SerialSchedulerImpl interface {
	NewSerialScheduler(conf *common.Configuration, connection *common.Connection,
		storage *backend.Storage, checkpointing bool) *SerialScheduler

	Run(application *application.Application) // &?

}

type SerialScheduler struct {

	// Configuration specifying node & system settings.
	configuration *common.Configuration

	// Connection for sending and receiving protocol messages.
	connection *common.Connection
	// Storage layer used in application execution.
	storage *backend.Storage

	// Should we check point?
	checkpointing bool

	SerialSchedulerImpl SerialSchedulerImpl
}

func (s *SerialScheduler) Close() error {
	return nil
}
