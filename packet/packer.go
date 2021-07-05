package packet

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/zhuangsirui/binpacker"
	"io"
)

//go:generate mockgen -destination mock/packer_mock.go -package mock . Packer

// MessageEntry is the unpacked message object.
type MessageEntry struct {
	ID   uint
	Data []byte
}

// Packer is a generic interface to pack and unpack message packet.
type Packer interface {
	// Pack packs Message into the packet to be written.
	// Pack(msg Message) ([]byte, error)
	Pack(entry *MessageEntry) ([]byte, error)

	// Unpack unpacks the message packet from reader,
	// returns the Message interface, and error if error occurred.
	Unpack(reader io.Reader) (*MessageEntry, error)
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
func (d *DefaultPacker) Pack(entry *MessageEntry) ([]byte, error) {
	size := len(entry.Data) // size without ID
	buff := bytes.NewBuffer(make([]byte, 0, size+4+4))

	p := binpacker.NewPacker(d.bytesOrder(), buff)
	if err := p.PushUint32(uint32(size)).Error(); err != nil {
		return nil, fmt.Errorf("write size err: %s", err)
	}
	if err := p.PushUint32(uint32(entry.ID)).Error(); err != nil {
		return nil, fmt.Errorf("write id err: %s", err)
	}
	if err := p.PushBytes(entry.Data).Error(); err != nil {
		return nil, fmt.Errorf("write data err: %s", err)
	}
	return buff.Bytes(), nil
}

// Unpack implements the Packer Unpack method.
func (d *DefaultPacker) Unpack(reader io.Reader) (*MessageEntry, error) {
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

	msg := &MessageEntry{
		ID:   uint(id),
		Data: data,
	}
	return msg, nil
}
