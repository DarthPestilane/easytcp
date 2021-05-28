package fixture

const ServerAddr = "127.0.0.1:8888"

// a group of message ids
const (
	_ uint = iota
	MsgIdPingReq
	MsgIdPingAck
)

// another group of message ids
const (
	_ uint = iota + 100
	MsgIdJson01Req
	MsgIdJson01Ack
)
