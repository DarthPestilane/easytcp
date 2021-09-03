package common

import (
	"encoding/binary"
	"fmt"
	"github.com/DarthPestilane/easytcp/message"
	"io"
)

type CustomPacker struct{}

func (p *CustomPacker) bytesOrder() binary.ByteOrder {
	return binary.BigEndian
}

func (p *CustomPacker) Pack(entry *message.Entry) ([]byte, error) {
	size := len(entry.Data) // without id
	buffer := make([]byte, 2+2+size)
	p.bytesOrder().PutUint16(buffer[:2], uint16(size))
	p.bytesOrder().PutUint16(buffer[2:4], entry.ID.(uint16))
	copy(buffer[4:], entry.Data)
	return buffer, nil
}

func (p *CustomPacker) Unpack(reader io.Reader) (*message.Entry, error) {
	headerBuffer := make([]byte, 2+2)
	if _, err := io.ReadFull(reader, headerBuffer); err != nil {
		return nil, fmt.Errorf("read size and id err: %s", err)
	}
	size := p.bytesOrder().Uint16(headerBuffer[:2])
	id := p.bytesOrder().Uint16(headerBuffer[2:])

	data := make([]byte, size)
	if _, err := io.ReadFull(reader, data); err != nil {
		return nil, fmt.Errorf("read data err: %s", err)
	}

	entry := &message.Entry{
		// since entry.ID is type of uint16, we need to use uint16 as well when adding routes.
		// eg: server.AddRoute(uint16(123), ...)
		ID:   id,
		Data: data,
	}
	entry.Set("theWholeLength", 2+2+size) // we can set our custom kv data here.
	// c.Message().Get("theWholeLength")  // and get them in route handler.
	return entry, nil
}
