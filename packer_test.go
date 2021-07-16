package easytcp

import (
	"bytes"
	"encoding/binary"
	"github.com/DarthPestilane/easytcp/message"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDefaultPacker(t *testing.T) {
	packer := &DefaultPacker{MaxSize: 1024}

	t.Run("when handle different types of id", func(t *testing.T) {
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
	})

	t.Run("when handle invalid type of id", func(t *testing.T) {
		entry := &message.Entry{
			ID:   "invalid",
			Data: []byte("test"),
		}
		msg, err := packer.Pack(entry)
		assert.Error(t, err)
		assert.Nil(t, msg)
	})

	t.Run("when size is too big", func(t *testing.T) {
		r := bytes.NewBuffer(nil)
		assert.NoError(t, binary.Write(r, binary.BigEndian, uint32(packer.MaxSize+1)))
		assert.NoError(t, binary.Write(r, binary.BigEndian, uint32(1)))
		assert.NoError(t, binary.Write(r, binary.BigEndian, []byte("test")))
		entry, err := packer.Unpack(r)
		assert.Error(t, err)
		assert.Nil(t, entry)
	})
}
