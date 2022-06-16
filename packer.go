package easytcp

import (
	"encoding/binary"
	"fmt"
	"github.com/spf13/cast"
	"io"
	"sync"
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
func NewDefaultPacker(maxDataSize ...int) *DefaultPacker {
	_maxDataSize := 1 << 10 << 10 // 1MB
	if len(maxDataSize) != 0 && maxDataSize[0] > 0 {
		_maxDataSize = maxDataSize[0]
	}
	return &DefaultPacker{
		MaxDataSize: _maxDataSize,
		buffPool: sync.Pool{
			New: func() interface{} {
				buff := make([]byte, 8+_maxDataSize)
				return &buff
			},
		},
		msgPool: sync.Pool{
			New: func() interface{} {
				return &Message{
					storage: make(map[string]interface{}),
				}
			},
		},
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

	buffPool sync.Pool // keeps *[]byte
	msgPool  sync.Pool // keeps *Message
}

func (d *DefaultPacker) bytesOrder() binary.ByteOrder {
	return binary.LittleEndian
}

// Pack implements the Packer Pack method.
func (d *DefaultPacker) Pack(msg *Message) ([]byte, error) {
	dataSize := len(msg.Data())
	if dataSize > d.MaxDataSize {
		return nil, fmt.Errorf("the dataSize %d is beyond the max: %d", dataSize, d.MaxDataSize)
	}

	buff := d.buffPool.Get().(*[]byte)
	defer d.buffPool.Put(buff)

	d.bytesOrder().PutUint32((*buff)[:4], uint32(dataSize)) // write dataSize
	id, err := cast.ToUint32E(msg.ID())
	if err != nil {
		return nil, fmt.Errorf("invalid type of msg.ID: %s", err)
	}
	d.bytesOrder().PutUint32((*buff)[4:8], id) // write id
	n := copy((*buff)[8:], msg.Data())         // write data
	return (*buff)[:8+n], nil
}

// Unpack implements the Packer Unpack method.
// Unpack returns the message whose ID is type of int.
// So we need use int id to register routes.
func (d *DefaultPacker) Unpack(reader io.Reader) (*Message, error) {
	buff := d.buffPool.Get().(*[]byte)
	defer d.buffPool.Put(buff)

	if _, err := io.ReadFull(reader, (*buff)[:8]); err != nil {
		return nil, fmt.Errorf("read size and id err: %s", err)
	}
	dataSize := d.bytesOrder().Uint32((*buff)[:4])
	if d.MaxDataSize > 0 && int(dataSize) > d.MaxDataSize {
		return nil, fmt.Errorf("the dataSize %d is beyond the max: %d", dataSize, d.MaxDataSize)
	}
	id := d.bytesOrder().Uint32((*buff)[4:8])
	if _, err := io.ReadFull(reader, (*buff)[8:8+dataSize]); err != nil {
		return nil, fmt.Errorf("read data err: %s", err)
	}

	// set the message
	msg := d.msgPool.Get().(*Message)
	defer d.msgPool.Put(msg)
	msg.Reset(int(id), (*buff)[8:8+dataSize])

	return msg, nil
}
