package packet

import (
	"encoding/binary"
	"fmt"
	"io"
)

type Packer interface {
	Pack()
	Unpack(reader io.Reader) (RawMessage, error) // 解包
}

// size|id|data
// 4|4|n
type DefaultPacker struct {
}

func (d *DefaultPacker) bytesOrder() binary.ByteOrder {
	return binary.LittleEndian
}

func (d *DefaultPacker) Pack() {
	panic("implement me")
}

func (d *DefaultPacker) Unpack(reader io.Reader) (RawMessage, error) {
	sizeBuff := make([]byte, 4)
	if _, err := io.ReadFull(reader, sizeBuff); err != nil {
		return nil, fmt.Errorf("read size err: %s", err)
	}
	size := d.bytesOrder().Uint32(sizeBuff)
	fmt.Println("size: ", size)

	idBuff := make([]byte, 4)
	if _, err := io.ReadFull(reader, idBuff); err != nil {
		return nil, fmt.Errorf("read id err: %s", err)

	}
	id := d.bytesOrder().Uint32(idBuff)

	data := make([]byte, size)
	if _, err := io.ReadFull(reader, data); err != nil {
		return nil, fmt.Errorf("read data err: %s", err)

	}

	msg := &DefaultRawMsg{
		Id:   id,
		Len:  size,
		Data: data,
	}
	return msg, nil
}
