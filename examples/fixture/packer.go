package fixture

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/DarthPestilane/easytcp/packet"
	"io"
)

type Msg16bit struct {
	Size uint16
	ID   uint16
	Data []byte
}

func (m *Msg16bit) Setup(id uint, data []byte) {
	m.ID = uint16(id)
	m.Data = data
	m.Size = uint16(len(data))
}

func (m *Msg16bit) Duplicate() packet.Message {
	return &Msg16bit{}
}

func (m *Msg16bit) GetID() uint {
	return uint(m.ID)
}

func (m *Msg16bit) GetSize() uint {
	return uint(m.Size)
}

func (m *Msg16bit) GetData() []byte {
	return m.Data
}

// Packer16bit custom packer
// packet format: size[2]id[2]data
type Packer16bit struct{}

func (p *Packer16bit) bytesOrder() binary.ByteOrder {
	return binary.BigEndian
}

func (p *Packer16bit) Pack(id uint, data []byte) ([]byte, error) {
	buff := bytes.NewBuffer(make([]byte, 0, len(data)+2+2))
	if err := binary.Write(buff, p.bytesOrder(), uint16(len(data))); err != nil {
		return nil, fmt.Errorf("write size err: %s", err)
	}
	if err := binary.Write(buff, p.bytesOrder(), uint16(id)); err != nil {
		return nil, fmt.Errorf("write id err: %s", err)
	}
	if err := binary.Write(buff, p.bytesOrder(), data); err != nil {
		return nil, fmt.Errorf("write data err: %s", err)
	}
	return buff.Bytes(), nil
}

func (p *Packer16bit) Unpack(reader io.Reader) (packet.Message, error) {
	sizeBuff := make([]byte, 2)
	if _, err := io.ReadFull(reader, sizeBuff); err != nil {
		return nil, fmt.Errorf("read size err: %s", err)
	}
	size := p.bytesOrder().Uint16(sizeBuff)

	idBuff := make([]byte, 2)
	if _, err := io.ReadFull(reader, idBuff); err != nil {
		return nil, fmt.Errorf("read id err: %s", err)
	}
	id := p.bytesOrder().Uint16(idBuff)

	data := make([]byte, size)
	if _, err := io.ReadFull(reader, data); err != nil {
		return nil, fmt.Errorf("read data err: %s", err)
	}

	msg := &Msg16bit{
		Size: size,
		ID:   id,
		Data: data,
	}
	return msg, nil
}

type JsonCodec struct {
}

func (c *JsonCodec) Encode(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

func (c *JsonCodec) Decode(data []byte, v interface{}) error {
	return json.Unmarshal(data, &v)
}
