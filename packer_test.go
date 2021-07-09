package easytcp

import (
	"bytes"
	"github.com/DarthPestilane/easytcp/message"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDefaultPacker(t *testing.T) {
	packer := &DefaultPacker{}
	ids := []interface{}{1, uint(1), uint32(1), uint64(1)}
	for _, id := range ids {
		entry := &message.Entry{
			ID:   id,
			Data: []byte("test"),
		}
		msg, err := packer.Pack(entry)
		assert.NoError(t, err)
		assert.NotNil(t, msg)

		r := bytes.NewBuffer(msg)
		newEntry, err := packer.Unpack(r)
		assert.NoError(t, err)
		assert.NotNil(t, newEntry)
		assert.EqualValues(t, newEntry.ID, entry.ID)
		assert.Equal(t, newEntry.Data, entry.Data)
	}

	// if id is a invalid type
	entry := &message.Entry{
		ID:   "invalid",
		Data: []byte("test"),
	}
	msg, err := packer.Pack(entry)
	assert.Error(t, err)
	assert.Nil(t, msg)
}
