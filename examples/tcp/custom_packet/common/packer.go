package common

import (
	"encoding/binary"
	"fmt"
	"github.com/DarthPestilane/easytcp/message"
	"github.com/spf13/cast"
	"io"
)

// CustomPacker treats packet as:
//
// totalSize(4)|idSize(2)|id(n)|data(n)
//
// | segment     | type   | size    | remark                |
// | ----------- | ------ | ------- | --------------------- |
// | `totalSize` | uint32 | 4       | the whole packet size |
// | `idSize`    | uint16 | 2       | length of id          |
// | `id`        | string | dynamic |                       |
// | `data`      | []byte | dynamic |                       |
type CustomPacker struct{}

func (p *CustomPacker) bytesOrder() binary.ByteOrder {
	return binary.LittleEndian
}

func (p *CustomPacker) Pack(entry *message.Entry) ([]byte, error) {
	// format: totalSize(4)|idSize(2)|id(n)|data(n)

	id, err := cast.ToStringE(entry.ID)
	if err != nil {
		return nil, err
	}
	buffer := make([]byte, 4+2+len(id)+len(entry.Data))
	p.bytesOrder().PutUint32(buffer[:4], uint32(len(buffer))) // write totalSize
	p.bytesOrder().PutUint16(buffer[4:6], uint16(len(id)))    // write idSize
	copy(buffer[6:6+len(id)], id)                             // write id
	copy(buffer[6+len(id):], entry.Data)                      // write data

	return buffer, nil
}

func (p *CustomPacker) Unpack(reader io.Reader) (*message.Entry, error) {
	// format: totalSize(4)|idSize(2)|id(n)|data(n)

	headerBuff := make([]byte, 4+2)
	if _, err := io.ReadFull(reader, headerBuff); err != nil {
		return nil, fmt.Errorf("read header err: %s", err)
	}
	totalSize := int(p.bytesOrder().Uint32(headerBuff[:4])) // read totalSize
	idSize := int(p.bytesOrder().Uint16(headerBuff[4:]))    // read idSize
	dataSize := totalSize - 4 - 2 - idSize

	bodyBuff := make([]byte, idSize+dataSize)
	if _, err := io.ReadFull(reader, bodyBuff); err != nil {
		return nil, fmt.Errorf("read body err: %s", err)
	}
	id := string(bodyBuff[:idSize]) // read id
	data := bodyBuff[idSize:]       // read body

	entry := &message.Entry{
		// ID is a string, so we should use a string-type id to register routes.
		// eg: server.AddRoute("string-id", handler)
		ID:   id,
		Data: data,
	}
	entry.Set("fullSize", totalSize)
	return entry, nil
}
