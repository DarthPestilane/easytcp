package easytcp

import (
	"encoding/binary"
	"fmt"
	"github.com/DarthPestilane/easytcp/message"
	"github.com/spf13/cast"
	"io"
)

//go:generate mockgen -destination internal/mock/packer_mock.go -package mock . Packer

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
// 		size: uint32 | took 4 bytes, only the size of `data`
// 		id:   uint32 | took 4 bytes
// 		data: []byte | took `size` bytes
type DefaultPacker struct {
	MaxSize int
}

func (d *DefaultPacker) bytesOrder() binary.ByteOrder {
	return binary.BigEndian
}

// Pack implements the Packer Pack method.
func (d *DefaultPacker) Pack(entry *message.Entry) ([]byte, error) {
	size := len(entry.Data) // only the size of `data`
	buffer := make([]byte, 4+4+size)
	d.bytesOrder().PutUint32(buffer[0:4], uint32(size)) // push size
	id, err := cast.ToUint32E(entry.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid type of entry.ID: %s", err)
	}
	d.bytesOrder().PutUint32(buffer[4:8], id) // push id
	copy(buffer[8:], entry.Data)              // push data
	return buffer, nil
}

// Unpack implements the Packer Unpack method.
func (d *DefaultPacker) Unpack(reader io.Reader) (*message.Entry, error) {
	headerBuffer := make([]byte, 4+4)
	if _, err := io.ReadFull(reader, headerBuffer); err != nil {
		return nil, fmt.Errorf("read size and id err: %s", err)
	}
	size := d.bytesOrder().Uint32(headerBuffer[0:4])
	if d.MaxSize > 0 && int(size) > d.MaxSize {
		return nil, fmt.Errorf("the size %d is beyond the max: %d", size, d.MaxSize)
	}
	id := d.bytesOrder().Uint32(headerBuffer[4:8])
	data := make([]byte, size)
	if _, err := io.ReadFull(reader, data); err != nil {
		return nil, fmt.Errorf("read data err: %s", err)
	}
	entry := &message.Entry{
		ID:   id,
		Data: data,
	}
	return entry, nil
}
