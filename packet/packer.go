package packet

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// Packer 打包/拆包
// 1. 对数据进行打包，得到消息
// 2. 对消息进行拆包，得到消息元数据
type Packer interface {
	Pack(id uint, data []byte) ([]byte, error) // 打包
	Unpack(reader io.Reader) (Message, error)  // 拆包
}

// DefaultPacker 默认的 Packer
// 包格式为:
//   size[4]id[4]data[n]
type DefaultPacker struct {
}

func (d *DefaultPacker) bytesOrder() binary.ByteOrder {
	return binary.LittleEndian
}

func (d *DefaultPacker) Pack(id uint, data []byte) ([]byte, error) {
	buff := bytes.NewBuffer([]byte{})
	size := len(data)
	if err := binary.Write(buff, d.bytesOrder(), uint32(size)); err != nil {
		return nil, fmt.Errorf("write size err: %s", err)
	}
	if err := binary.Write(buff, d.bytesOrder(), uint32(id)); err != nil {
		return nil, fmt.Errorf("write id err: %s", err)
	}
	if err := binary.Write(buff, d.bytesOrder(), data); err != nil {
		return nil, fmt.Errorf("write data err: %s", err)
	}
	return buff.Bytes(), nil
}

func (d *DefaultPacker) Unpack(reader io.Reader) (Message, error) {
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

	msg := &DefaultMsg{
		Id:   id,
		Size: size,
		Data: data,
	}
	return msg, nil
}
