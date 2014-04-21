package resolver_rpc

// Status represents the status of a RPC's reply.
type Status int

const (
	OK           Status = iota + 1 // The RPC was a success.
	KeyNotFound                    // The specified key does not exist.
	ItemNotFound                   // The specified item does not exist.
	WrongServer                    // The specified key does not fall in the server's hash range.
	ItemExists                     // The item already exists in the list.
	NotReady                       // The storage servers are still getting ready.
)

type Node struct {
	HostPort string // The host:port address of the storage server node.
	NodeID   uint32 // The ID identifying this storage server node.
}

type RegisterArgs struct {
	ServerInfo Node
}

type RegisterReply struct {
	Status  Status
	Servers []Node
}

type GetServersArgs struct {
	// Intentionally left empty.
}

type GetServersReply struct {
	Status  Status
	Servers []Node
}

type GetArgs struct {
	Key       string
	WantLease bool
	HostPort  string // The Libstore's callback host:port.
}

type GetReply struct {
	Status Status
	Value  string
}

type GetListReply struct {
	Status Status
	Value  []string
}

type PutArgs struct {
	Key   string
	Value string
}

type PutReply struct {
	Status Status
}
