package resolver_rpc

type TradeResolver interface {
	RegisterServer(*RegisterArgs, *RegisterReply) error
	GetServers(*GetServersArgs, *GetServersReply) error
}

type Resolver struct {
	TradeResolver
}

func Wrap(s TradeResolver) *Resolver {
	return &Resolver{s}
}
