package packet

// Message 对原始消息的抽象
// 通常由 Packer.Unpack() 拆包得到
type Message interface {
	GetId() uint
	GetSize() uint

	// GetData 返回原始消息中的 data 部分
	// 通常将经过 Codec.Decode() 处理
	GetData() []byte
}

type DefaultMsg struct {
	Id   uint32
	Size uint32
	Data []byte
}

func (d *DefaultMsg) GetId() uint {
	return uint(d.Id)
}

func (d *DefaultMsg) GetSize() uint {
	return uint(d.Size)
}

func (d *DefaultMsg) GetData() []byte {
	return d.Data
}
