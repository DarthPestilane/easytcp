package fixture

const ServerAddr = "0.0.0.0:8888"

// a group of message ids
const (
	_ uint32 = iota
	MsgIdPingReq
	MsgIdPingAck
)

// broadcast messages
const (
	_ uint32 = iota + 200
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
