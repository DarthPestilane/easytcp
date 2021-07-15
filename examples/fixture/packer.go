package fixture

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/DarthPestilane/easytcp/message"
	"io"
)

// CustomPacker custom packer
// payload format:
// totalSize(4)|typeSize(2)|type(n)|data(n)
//
// | `totalSize` | `typeSize` | `type`                  | `data` |
// | ----------- | ---------- | ----------------------- | ------ |
// | uint32      | uint16     | []byte                  | []byte |
// | 4 bytes     | 2 bytes    | according to `typeSize` |        |
type CustomPacker struct{}

func (p *CustomPacker) bytesOrder() binary.ByteOrder {
	return binary.LittleEndian
}

func (p *CustomPacker) Pack(entry *message.Entry) ([]byte, error) {
	// format: totalSize(4)|typeSize(2)|type(n)|data(n)
	var typ string
	switch v := entry.ID.(type) {
	case string:
		typ = v
	case []byte:
		typ = string(v)
	case fmt.Stringer:
		typ = v.String()
	default:
		return nil, fmt.Errorf("invalid type of entry.ID: %T", entry.ID)
	}
	typeSizeLen := 2
	typeSizeVal := len(typ)

	totalSizeLen := 4
	totalSizeVal := typeSizeLen + typeSizeVal + len(entry.Data)

	bufferLen := totalSizeLen + totalSizeVal
	buffer := bytes.NewBuffer(make([]byte, 0, bufferLen))

	// write totalSize
	if err := binary.Write(buffer, p.bytesOrder(), uint32(totalSizeVal)); err != nil {
		return nil, fmt.Errorf("write totalSize err: %s", err)
	}

	// write typeSize
	if err := binary.Write(buffer, p.bytesOrder(), uint16(typeSizeVal)); err != nil {
		return nil, fmt.Errorf("write typeSize err: %s", err)
	}

	// write type
	if err := binary.Write(buffer, p.bytesOrder(), []byte(typ)); err != nil {
		return nil, fmt.Errorf("write type/id err: %s", err)
	}

	// write data
	if err := binary.Write(buffer, p.bytesOrder(), entry.Data); err != nil {
		return nil, fmt.Errorf("write data err: %s", err)
	}

	return buffer.Bytes(), nil
}

func (p *CustomPacker) Unpack(reader io.Reader) (*message.Entry, error) {
	// format: totalSize(4)|typeSize(2)|type(n)|data(n)

	// read totalSize
	totalSizeBuff := make([]byte, 4)
	if _, err := io.ReadFull(reader, totalSizeBuff); err != nil {
		return nil, fmt.Errorf("read totalSize err: %s", err)
	}
	totalSizeVal := p.bytesOrder().Uint32(totalSizeBuff)

	// read typeSize
	typeSizeBuff := make([]byte, 2)
	if _, err := io.ReadFull(reader, typeSizeBuff); err != nil {
		return nil, fmt.Errorf("read typeSize err: %s", err)
	}
	typeSizeVal := p.bytesOrder().Uint16(typeSizeBuff)

	// read type
	typeBuff := make([]byte, typeSizeVal)
	if _, err := io.ReadFull(reader, typeBuff); err != nil {
		return nil, fmt.Errorf("read type/id err: %s", err)
	}

	// read data
	dataBuff := make([]byte, uint(totalSizeVal)-2-uint(typeSizeVal))
	if _, err := io.ReadFull(reader, dataBuff); err != nil {
		return nil, fmt.Errorf("read data err: %s", err)
	}

	entry := &message.Entry{
		// ID is a string, so we should use a string-type id to register routes.
		// eg: server.AddRoute("string-id", handler)
		ID:   string(typeBuff),
		Data: dataBuff,
	}
	return entry, nil
}
