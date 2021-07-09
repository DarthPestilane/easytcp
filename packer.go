package easytcp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/DarthPestilane/easytcp/message"
	"github.com/zhuangsirui/binpacker"
	"io"
)

//go:generate mockgen -destination mock/packer_mock.go -package mock . Packer

// Packer is a generic interface to pack and unpack message packet.
type Packer interface {
	// Pack packs Message into the packet to be written.
	// Pack(msg Message) ([]byte, error)
	Pack(entry *message.Entry) ([]byte, error)

	// Unpack unpacks the message packet from reader,
	// returns the Message interface, and error if error occurred.
	Unpack(reader io.Reader) (*message.Entry, error)
}

var _ Packer = &DefaultPacker{}

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

func (d *DefaultPacker) assertID(id interface{}) (uint32, bool) {
	switch v := id.(type) {
	case uint:
		return uint32(v), true
	case uint32:
		return v, true
	case uint64:
		return uint32(v), true
	case int:
		return uint32(v), true
	default:
		return 0, false
	}
}

// Pack implements the Packer Pack method.
func (d *DefaultPacker) Pack(entry *message.Entry) ([]byte, error) {
	size := len(entry.Data) // size without ID
	buff := bytes.NewBuffer(make([]byte, 0, size+4+4))

	p := binpacker.NewPacker(d.bytesOrder(), buff)
	if err := p.PushUint32(uint32(size)).Error(); err != nil {
		return nil, fmt.Errorf("write size err: %s", err)
	}
	id, ok := d.assertID(entry.ID)
	if !ok {
		return nil, fmt.Errorf("invalid type of entry.ID: %v(%T)", entry.ID, entry.ID)
	}
	if err := p.PushUint32(id).Error(); err != nil {
		return nil, fmt.Errorf("write id err: %s", err)
	}
	if err := p.PushBytes(entry.Data).Error(); err != nil {
		return nil, fmt.Errorf("write data err: %s", err)
	}
	return buff.Bytes(), nil
}

// Unpack implements the Packer Unpack method.
func (d *DefaultPacker) Unpack(reader io.Reader) (*message.Entry, error) {
	// We should use io.ReadFull method, especially called conn.SetReadBuffer(n).

	sizeBuff := make([]byte, 4)
	if _, err := io.ReadFull(reader, sizeBuff); err != nil {
		return nil, fmt.Errorf("read size err: %s", err)
	}
	size := d.bytesOrder().Uint32(sizeBuff)

	idBuff := make([]byte, 4)
	if _, err := io.ReadFull(reader, idBuff); err != nil {
		return nil, fmt.Errorf("read id err: %s", err)
	}
	id := d.bytesOrder().Uint32(idBuff)

	data := make([]byte, size)
	if _, err := io.ReadFull(reader, data); err != nil {
		return nil, fmt.Errorf("read data err: %s", err)
	}

	msg := &message.Entry{
		ID:   id,
		Data: data,
	}
	return msg, nil
}
