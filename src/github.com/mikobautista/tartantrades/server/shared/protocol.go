package shared

type MessageType int

const (
	AUTHENTICATION MessageType = iota
)

type TradeMessage struct {
	Type    MessageType
	Payload string
}
