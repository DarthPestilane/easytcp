package common

import (
	"bytes"
	"github.com/DarthPestilane/easytcp"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCustomPacker(t *testing.T) {
	packer := &CustomPacker{}
	msg := easytcp.NewMessage("test", []byte("data"))
	packedBytes, err := packer.Pack(msg)
	assert.NoError(t, err)
	assert.NotNil(t, packedBytes)

	r := bytes.NewBuffer(packedBytes)
	newMsg, err := packer.Unpack(r)
	assert.NoError(t, err)
	assert.NotNil(t, newMsg)
	assert.Equal(t, newMsg.ID(), msg.ID())
	assert.Equal(t, newMsg.Data(), msg.Data())
}
