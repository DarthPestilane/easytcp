package fixture

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/DarthPestilane/easytcp/message"
	"io"
)

// Packer16bit custom packer
// packet format: size[2]id[2]data
type Packer16bit struct{}

func (p *Packer16bit) bytesOrder() binary.ByteOrder {
	return binary.BigEndian
}

func (p *Packer16bit) Pack(msg *message.Entry) ([]byte, error) {
	size := len(msg.Data) // without id
	buff := bytes.NewBuffer(make([]byte, 0, size+2+2))
	if err := binary.Write(buff, p.bytesOrder(), uint16(size)); err != nil {
		return nil, fmt.Errorf("write size err: %s", err)
	}
	if err := binary.Write(buff, p.bytesOrder(), uint16(msg.ID)); err != nil {
		return nil, fmt.Errorf("write id err: %s", err)
	}
	if err := binary.Write(buff, p.bytesOrder(), msg.Data); err != nil {
		return nil, fmt.Errorf("write data err: %s", err)
	}
	return buff.Bytes(), nil
}

func (p *Packer16bit) Unpack(reader io.Reader) (*message.Entry, error) {
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

	msg := &message.Entry{ID: uint(id), Data: data}
	return msg, nil
}
