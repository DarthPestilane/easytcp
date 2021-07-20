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

// NewDefaultPacker create a *DefaultPacker with initial field value.
func NewDefaultPacker() *DefaultPacker {
	return &DefaultPacker{MaxSize: 1024 * 1024}
}

// DefaultPacker is the default Packer used in session.
// DefaultPacker treats the packet with the format:
// 	(size)(id)(data):
// 		size: uint32 | took 4 bytes, only the size of data
// 		id: uint32   | took 4 bytes
// 		data: []byte | length is the size
type DefaultPacker struct {
	MaxSize int
}

func (d *DefaultPacker) bytesOrder() binary.ByteOrder {
	return binary.BigEndian
}

func (d *DefaultPacker) assertID(id interface{}) (uint32, bool) {
	switch v := id.(type) {
	case uint32:
		return v, true
	case *uint32:
		return *v, true
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
		theErr := fmt.Errorf("read size err: %s", err)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil, &UnpackError{Err: theErr}
		}
		return nil, theErr
	}
	size := d.bytesOrder().Uint32(sizeBuff)

	if d.MaxSize > 0 && int(size) > d.MaxSize {
		return nil, fmt.Errorf("the size %d is beyond the max: %d", size, d.MaxSize)
	}

	idBuff := make([]byte, 4)
	if _, err := io.ReadFull(reader, idBuff); err != nil {
		return nil, &UnpackError{Err: fmt.Errorf("read id err: %s", err)}
	}
	id := d.bytesOrder().Uint32(idBuff)

	data := make([]byte, size)
	if _, err := io.ReadFull(reader, data); err != nil {
		return nil, &UnpackError{Err: fmt.Errorf("read data err: %s", err)}
	}

	msg := &message.Entry{
		ID:   id,
		Data: data,
	}
	return msg, nil
}
