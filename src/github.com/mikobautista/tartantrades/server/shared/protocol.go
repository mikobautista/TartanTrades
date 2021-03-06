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
	// AcceptedId holds current index
	RECOVER_CHECK
	RECOVER_NECESSARY

	// Paxos
	PREPARE
	PROMISE
	NPROMISE
	COMMIT
	ACCEPT
	ACCEPTED
)

type ResolverMessage struct {
	Type    ResolverMessageType
	Payload string
	Id      uint32
}

type TransactionType int

const (
	PURCHASE TransactionType = iota
	SELL
)

type Transaction struct {
	Type TransactionType
	// For purchasing
	Commit uint32
	To     uint32

	// For Selling
	X  string
	Y  string
	Id uint32
}

type TradeMessage struct {
	Type             TradeMessageType
	Payload          string
	Recovery_Payload string
	FromNodeId       uint32

	ProposedId uint32
	PromisedId uint32

	AcceptedId    uint32
	AcceptedValue Transaction
}
