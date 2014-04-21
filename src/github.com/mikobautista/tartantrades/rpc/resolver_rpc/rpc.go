package resolver_rpc

// STAFF USE ONLY! Students should not use this interface in their code.
type RemoteStorageServer interface {
	RegisterServer(*RegisterArgs, *RegisterReply) error
	GetServers(*GetServersArgs, *GetServersReply) error
	Get(*GetArgs, *GetReply) error
	GetList(*GetArgs, *GetListReply) error
	Put(*PutArgs, *PutReply) error
	AppendToList(*PutArgs, *PutReply) error
	RemoveFromList(*PutArgs, *PutReply) error
}

type StorageServer struct {
	RemoteStorageServer
}

func Wrap(s RemoteStorageServer) RemoteStorageServer {
	return &StorageServer{s}
}
