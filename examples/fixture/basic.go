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

// broadcast messages
const (
	_ uint = iota + 200
	MsgIdBroadCastReq
	MsgIdBroadCastAck
)

type Json01Req struct {
	Key1 string `json:"key_1"`
	Key2 int    `json:"key_2"`
	Key3 bool   `json:"key_3"`
}

type Json01Resp struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}
