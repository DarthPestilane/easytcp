package packet

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/zhuangsirui/binpacker"
	"testing"
)

func TestDefaultPacker_Pack(t *testing.T) {
	p := &DefaultPacker{}
	id := uint32(123)
	data := []byte("hello")
	size := uint32(len(data))
	msg, err := p.Pack(uint(id), data)
	assert.NoError(t, err)

	unpacker := binpacker.NewUnpacker(p.bytesOrder(), bytes.NewReader(msg))
	size2, err := unpacker.ShiftUint32()
	assert.NoError(t, err)
	assert.Equal(t, size, size2)

	id2, err := unpacker.ShiftUint32()
	assert.NoError(t, err)
	assert.Equal(t, id, id2)

	data2, err := unpacker.ShiftBytes(uint64(size))
	assert.NoError(t, err)
	assert.Equal(t, data, data2)
}

func TestDefaultPacker_Unpack(t *testing.T) {
	id := uint32(123)
	data := []byte("hello")
	size := uint32(len(data))

	p := &DefaultPacker{}
	buff := bytes.NewBuffer(nil)
	packer := binpacker.NewPacker(p.bytesOrder(), buff)
	err := packer.PushUint32(size).PushUint32(id).PushBytes(data).Error()
	assert.NoError(t, err)

	msg, err := p.Unpack(buff)
	assert.NoError(t, err)
	assert.IsType(t, &DefaultMsg{}, msg)
	assert.EqualValues(t, msg.GetSize(), size)
	assert.EqualValues(t, msg.GetID(), id)
	assert.Equal(t, msg.GetData(), data)
}
