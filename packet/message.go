package packet

// Message 对原始消息的抽象
// 通常由 Packer.Unpack() 拆包得到
type Message interface {
	GetId() uint32
	GetLen() uint32

	// GetData 返回原始消息中的 data 部分
	// 通常将经过 Codec.Decode() 处理
	GetData() []byte
}

type DefaultRawMsg struct {
	Id   uint32
	Len  uint32
	Data []byte
}

func (d *DefaultRawMsg) GetId() uint32 {
	return d.Id
}

func (d *DefaultRawMsg) GetLen() uint32 {
	return d.Len
}

func (d *DefaultRawMsg) GetData() []byte {
	return d.Data
}
