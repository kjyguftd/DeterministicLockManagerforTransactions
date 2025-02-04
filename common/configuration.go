package common

type Node struct {
	// Globally unique node identifier.
	NodeID      int
	ReplicaID   int
	PartitionID int

	// IP address of this node's machine.
	Host string

	// Port on which to listen for messages from other nodes.
	Port int

	// Total number of cores available for use by this node.
	// Note: Is this needed?
	Cores int
}

type ConfigurationImpl interface {
	// LookupPartition Returns the node_id of the partition at which 'key' is stored.
	LookupPartition(key Key) int

	// WriteToFile Dump the current config into the file in key=value format.
	// Returns true when success.
	WriteToFile(filename string) bool

	// TODO:Comments
	processConfigLine(key string, value string)
	readFromFile(filename string) int
}

type Configuration struct {
	// This node's node_id.
	NodeID int

	Filename string

	// Tracks the set of current active nodes in the system.
	Nodes map[int]*Node

	ConfigImpl ConfigurationImpl
}

func NewConfiguration(nodeID int, filename string, configImpl ConfigurationImpl) *Configuration {
	return &Configuration{NodeID: nodeID, Filename: filename, ConfigImpl: configImpl}
}
