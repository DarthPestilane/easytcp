package easytcp

import (
	"encoding/binary"
	"fmt"
	"github.com/spf13/cast"
	"io"
)

//go:generate mockgen -destination ./packer_mock.go -package easytcp . Packer

// Packer is a generic interface to pack and unpack message packet.
type Packer interface {
	// Pack packs Message into the packet to be written.
	Pack(msg *Message) ([]byte, error)

	// Unpack unpacks the message packet from reader,
	// returns the message, and error if error occurred.
	Unpack(reader io.Reader) (*Message, error)
}

var _ Packer = &DefaultPacker{}

// NewDefaultPacker create a *DefaultPacker with initial field value.
func NewDefaultPacker() *DefaultPacker {
	return &DefaultPacker{
		MaxDataSize: 1 << 10 << 10, // 1MB
	}
}

// DefaultPacker is the default Packer used in session.
// Treats the packet with the format:
//
// dataSize(4)|id(4)|data(n)
//
// | segment    | type   | size    | remark                  |
// | ---------- | ------ | ------- | ----------------------- |
// | `dataSize` | uint32 | 4       | the size of `data` only |
// | `id`       | uint32 | 4       |                         |
// | `data`     | []byte | dynamic |                         |
// .
type DefaultPacker struct {
	// MaxDataSize represents the max size of `data`
	MaxDataSize int
}

func (d *DefaultPacker) bytesOrder() binary.ByteOrder {
	return binary.LittleEndian
}

// Pack implements the Packer Pack method.
func (d *DefaultPacker) Pack(msg *Message) ([]byte, error) {
	dataSize := len(msg.Data())
	if d.MaxDataSize > 0 && dataSize > d.MaxDataSize {
		return nil, fmt.Errorf("the dataSize %d is beyond the max: %d", dataSize, d.MaxDataSize)
	}
	buffer := make([]byte, 4+4+dataSize)
	d.bytesOrder().PutUint32(buffer[:4], uint32(dataSize)) // write dataSize
	id, err := cast.ToUint32E(msg.ID())
	if err != nil {
		return nil, fmt.Errorf("invalid type of msg.ID: %s", err)
	}
	d.bytesOrder().PutUint32(buffer[4:8], id) // write id
	copy(buffer[8:], msg.Data())              // write data
	return buffer, nil
}

// Unpack implements the Packer Unpack method.
// Unpack returns the message whose ID is type of int.
// So we need use int id to register routes.
func (d *DefaultPacker) Unpack(reader io.Reader) (*Message, error) {
	headerBuffer := make([]byte, 4+4)
	if _, err := io.ReadFull(reader, headerBuffer); err != nil {
		if err == io.EOF {
			return nil, err
		}
		return nil, fmt.Errorf("read size and id err: %s", err)
	}
	dataSize := d.bytesOrder().Uint32(headerBuffer[:4])
	if d.MaxDataSize > 0 && int(dataSize) > d.MaxDataSize {
		return nil, fmt.Errorf("the dataSize %d is beyond the max: %d", dataSize, d.MaxDataSize)
	}
	id := d.bytesOrder().Uint32(headerBuffer[4:8])
	data := make([]byte, dataSize)
	if _, err := io.ReadFull(reader, data); err != nil {
		if err == io.EOF {
			return nil, err
		}
		return nil, fmt.Errorf("read data err: %s", err)
	}
	return NewMessage(int(id), data), nil
}
