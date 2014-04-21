package resolver_rpc

// STAFF USE ONLY! Students should not use this interface in their code.
type TradeResolver interface {
	RegisterServer(*RegisterArgs, *RegisterReply) error
	GetServers(*GetServersArgs, *GetServersReply) error
	Get(*GetArgs, *GetReply) error
	GetList(*GetArgs, *GetListReply) error
	Put(*PutArgs, *PutReply) error
	AppendToList(*PutArgs, *PutReply) error
	RemoveFromList(*PutArgs, *PutReply) error
}

type Resolver struct {
	TradeResolver
}

func Wrap(s TradeResolver) *Resolver {
	return &Resolver{s}
}
