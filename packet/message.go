package packet

// Message 对原始消息的抽象
// 通常由 Packer.Unpack() 拆包得到
type Message interface {
	GetSize() uint   // 返回消息的 size 长度部分
	GetId() uint     // 返回消息的 id 标识部分, 用于消息路由
	GetData() []byte // 返回消息的 data 数据部分, 通常将经过 Codec.Decode() 处理
}

var _ Message = &DefaultMsg{}

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
