package packet

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/zhuangsirui/binpacker"
	"testing"
)

func TestDefaultPacker_Pack(t *testing.T) {
	id := uint(123)
	data := []byte("hello")
	size := uint32(len(data))
	rawMsg := &MessageEntry{
		ID:   id,
		Data: data,
	}

	p := &DefaultPacker{}
	packedMsg, err := p.Pack(rawMsg)
	assert.NoError(t, err)

	unpacker := binpacker.NewUnpacker(p.bytesOrder(), bytes.NewReader(packedMsg))
	size2, err := unpacker.ShiftUint32()
	assert.NoError(t, err)
	assert.Equal(t, size, size2)

	id2, err := unpacker.ShiftUint32()
	assert.NoError(t, err)
	assert.EqualValues(t, id, id2)

	data2, err := unpacker.ShiftBytes(uint64(size))
	assert.NoError(t, err)
	assert.Equal(t, data, data2)
}

func TestDefaultPacker_Unpack(t *testing.T) {
	id := uint(123)
	data := []byte("hello")
	size := len(data)

	p := &DefaultPacker{}
	buff := bytes.NewBuffer(nil)
	packer := binpacker.NewPacker(p.bytesOrder(), buff)
	err := packer.PushUint32(uint32(size)).PushUint32(uint32(id)).PushBytes(data).Error()
	assert.NoError(t, err)

	msg, err := p.Unpack(buff)
	assert.NoError(t, err)
	assert.IsType(t, &MessageEntry{}, msg)
	assert.Len(t, msg.Data, size)
	assert.EqualValues(t, msg.ID, id)
	assert.Equal(t, msg.Data, data)
}
