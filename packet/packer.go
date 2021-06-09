package packet

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/zhuangsirui/binpacker"
	"io"
)

//go:generate mockgen -destination mock/packer_mock.go -package mock . Packer

// Packer 打包/拆包
// 1. 对数据进行打包，得到消息
// 2. 对消息进行拆包，得到消息元数据.
type Packer interface {
	Pack(id uint, data []byte) ([]byte, error) // 打包
	Unpack(reader io.Reader) (Message, error)  // 拆包
}

// DefaultPacker 默认的 Packer
// 封包格式:
//   size[4]id[4]data.
type DefaultPacker struct{}

func (d *DefaultPacker) bytesOrder() binary.ByteOrder {
	return binary.LittleEndian
}

func (d *DefaultPacker) Pack(id uint, data []byte) ([]byte, error) {
	buff := bytes.NewBuffer(make([]byte, 0, len(data)+4+4))
	p := binpacker.NewPacker(d.bytesOrder(), buff)
	if err := p.PushUint32(uint32(len(data))).Error(); err != nil {
		return nil, fmt.Errorf("write size err: %s", err)
	}
	if err := p.PushUint32(uint32(id)).Error(); err != nil {
		return nil, fmt.Errorf("write id err: %s", err)
	}
	if err := p.PushBytes(data).Error(); err != nil {
		return nil, fmt.Errorf("write data err: %s", err)
	}
	return buff.Bytes(), nil
}

func (d *DefaultPacker) Unpack(reader io.Reader) (Message, error) {
	p := binpacker.NewUnpacker(d.bytesOrder(), reader)
	size, err := p.ShiftUint32()
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("read size err: %s", err)
	}
	id, err := p.ShiftUint32()
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("read id err: %s", err)
	}
	data, err := p.ShiftBytes(uint64(size))
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("read data err: %s", err)
	}
	msg := &DefaultMsg{
		ID:   id,
		Size: size,
		Data: data,
	}
	return msg, nil
}
