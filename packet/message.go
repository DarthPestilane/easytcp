package packet

type RawMessage interface {
	GetId() uint32
	GetLen() uint32
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
