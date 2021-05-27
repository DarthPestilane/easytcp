package packet

// Request 请求
// 当 session.Session 读取到消息后，经过 Packer.Unpack() 和 Codec.Decode() 后
// 构建出 Request，发送到 channel 中，等待 router.Router 消费
type Request struct {
	// Id 消息ID
	Id uint32

	// Data 从原始消息中，拆包、解码后得到的数据，通常需要手动断言处理
	Data interface{}
}

// Response 响应
// 通常随路由 handler (router.HandleFunc) 的返回，
// 并发送到 session.Session 的 channel 中，在 SendResp 方法里被消费。
// Data 会由 Codec.Encode() 进行编码后，和 Id 一起由 Packer.Pack() 封包成最终待发送的消息
type Response struct {
	// Id 消息ID
	Id uint32

	// Data 未经过编码的数据
	Data interface{}
}
