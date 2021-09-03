package common

import (
	"bytes"
	"github.com/DarthPestilane/easytcp/message"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCustomPacker(t *testing.T) {
	packer := &CustomPacker{}

	entry := &message.Entry{
		ID:   uint16(123),
		Data: []byte("data"),
	}
	msg, err := packer.Pack(entry)
	assert.NoError(t, err)
	assert.NotNil(t, msg)

	r := bytes.NewBuffer(msg)
	newEntry, err := packer.Unpack(r)
	assert.NoError(t, err)
	assert.NotNil(t, newEntry)
	assert.Equal(t, newEntry.ID, entry.ID)
	assert.Equal(t, newEntry.Data, entry.Data)
}
