package common

import (
	"encoding/binary"
	"fmt"
	"github.com/DarthPestilane/easytcp"
	"io"
)

// CustomPacker treats packet as:
//
// totalSize(4)|id(4)|data(n)
//
// | segment     | type   | size    | remark                |
// | ----------- | ------ | ------- | --------------------- |
// | `totalSize` | uint32 | 4       | the whole packet size |
// | `id`        | uint32 | 4       |                       |
// | `data`      | []byte | dynamic |                       |
type CustomPacker struct{}

func (p *CustomPacker) Pack(msg *easytcp.Message) ([]byte, error) {
	buffer := make([]byte, 4+4+len(msg.Data()))
	p.byteOrder().PutUint32(buffer[0:4], uint32(len(buffer)))   // write totalSize
	p.byteOrder().PutUint32(buffer[4:8], uint32(msg.ID().(ID))) // write id
	copy(buffer[8:], msg.Data())                                // write data
	return buffer, nil
}

func (p *CustomPacker) Unpack(reader io.Reader) (*easytcp.Message, error) {
	headerBuffer := make([]byte, 4+4)
	if _, err := io.ReadFull(reader, headerBuffer); err != nil {
		if err == io.EOF {
			return nil, err
		}
		return nil, fmt.Errorf("read header from reader err: %s", err)
	}
	totalSize := p.byteOrder().Uint32(headerBuffer[:4]) // read totalSize
	id := ID(p.byteOrder().Uint32(headerBuffer[4:]))    // read id

	// read data
	dataSize := totalSize - 4 - 4
	data := make([]byte, dataSize)
	if _, err := io.ReadFull(reader, data); err != nil {
		return nil, fmt.Errorf("read data from reader err: %s", err)
	}
	return easytcp.NewMessage(id, data), nil
}

func (*CustomPacker) byteOrder() binary.ByteOrder {
	return binary.BigEndian
}
