package shared

type ResolverMessageType int
type TradeMessageType int

const (
	// Sent from trade server to resolver
	// Token in payload field
	AUTHENTICATION ResolverMessageType = iota

	// Sent from resolver to trade servers
	// trade tcp host:port in payload field
	TRADE_SERVER_JOIN
	TRADE_SERVER_DROP

	// trade id in id field
	ID_ASSIGNMENT

	CONNECT
)

const (
	WELCOME TradeMessageType = iota
	PREPARE
	COMMIT
	ACCEPT
)

type ResolverMessage struct {
	Type    ResolverMessageType
	Payload string
	Id      uint32
}

type TradeMessage struct {
	Type    TradeMessageType
	Payload string
	Id      uint32
}
