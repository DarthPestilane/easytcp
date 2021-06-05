package packet

// Request 请求
// 当 session.Session 读取到消息后，经过 Packer.Unpack() 和 Codec.Decode() 后
// 构建出 Request，发送到 channel 中，等待 router.Router 消费
type Request struct {
	Id      uint   // 消息ID
	RawSize uint   // 原始消息的长度
	RawData []byte // 原始消息中的数据段
}

// Response 响应
// 通常随路由 handler (router.HandleFunc) 的返回，
// 并发送到 session.Session 的 channel 中，在 SendResp 方法里被消费。
// Data 会由 Codec.Encode() 进行编码后，和 Id 一起由 Packer.Pack() 封包成最终待发送的消息
type Response struct {
	Id   uint        // 消息ID
	Data interface{} // 未经过编码的数据
}
