package packet

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/zhuangsirui/binpacker"
	"io"
)

//go:generate mockgen -destination mock/packer_mock.go -package mock . Packer

// Packer is a generic interface to pack and unpack message packet.
type Packer interface {
	// Pack packs data into the message to be sent.
	Pack(id uint, data []byte) ([]byte, error)

	// Unpack unpacks the message packet from reader,
	// returns the Message interface, and error if error occurred.
	Unpack(reader io.Reader) (Message, error)
}

// DefaultPacker is the default Packer used in session.
// DefaultPacker treats the packet with the format:
// 	(size)(id)(data):
// 		size: uint32 | took 4 bytes
// 		id: uint32   | took 4 bytes
// 		data: []byte | length is the size
type DefaultPacker struct{}

func (d *DefaultPacker) bytesOrder() binary.ByteOrder {
	return binary.BigEndian
}

// Pack implements the Packer Pack method.
func (d *DefaultPacker) Pack(id uint, data []byte) ([]byte, error) {
	buff := bytes.NewBuffer(make([]byte, 0, len(data)+4+4))
	p := binpacker.NewPacker(d.bytesOrder(), buff)
	if err := p.PushUint32(uint32(len(data))).Error(); err != nil {
		return nil, fmt.Errorf("write size err: %s", err)
	}
	if err := p.PushUint32(uint32(id)).Error(); err != nil {
		return nil, fmt.Errorf("write id err: %s", err)
	}
	if err := p.PushBytes(data).Error(); err != nil {
		return nil, fmt.Errorf("write data err: %s", err)
	}
	return buff.Bytes(), nil
}

// Unpack implements the Packer Unpack method.
func (d *DefaultPacker) Unpack(reader io.Reader) (Message, error) {
	p := binpacker.NewUnpacker(d.bytesOrder(), reader)
	size, err := p.ShiftUint32()
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("read size err: %s", err)
	}
	id, err := p.ShiftUint32()
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("read id err: %s", err)
	}
	data, err := p.ShiftBytes(uint64(size))
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("read data err: %s", err)
	}
	msg := &DefaultMsg{
		ID:   id,
		Size: size,
		Data: data,
	}
	return msg, nil
}
