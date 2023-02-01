package easytcp

import (
	"bytes"
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestDefaultPacker_PackAndUnpack(t *testing.T) {
	packer := NewDefaultPacker()

	t.Run("when handle different types of id", func(t *testing.T) {
		var testIdInt = 1
		var testIdInt32 int32 = 1
		var testIdInt64 int64 = 1

		var testIdUint uint = 1
		var testIdUint32 uint32 = 1
		var testIdUint64 uint64 = 1

		ids := []interface{}{
			testIdInt, &testIdInt,
			testIdInt32, &testIdInt32,
			testIdInt64, &testIdInt64,

			testIdUint, &testIdUint,
			testIdUint32, &testIdUint32,
			testIdUint64, &testIdUint64,
		}
		for _, id := range ids {
			msg := NewMessage(id, []byte("test"))
			packedBytes, err := packer.Pack(msg)
			assert.NoError(t, err)
			assert.NotNil(t, packedBytes)
			assert.Equal(t, packedBytes[8:], []byte("test"))

			r := bytes.NewBuffer(packedBytes)
			newMsg, err := packer.Unpack(r)
			assert.NoError(t, err)
			assert.NotNil(t, newMsg)
			assert.EqualValues(t, reflect.Indirect(reflect.ValueOf(msg.ID())).Interface(), newMsg.ID())
			assert.Equal(t, newMsg.Data(), msg.Data())
		}
	})

	t.Run("when handle invalid type of id", func(t *testing.T) {
		msg := NewMessage("cannot cast to uint32", []byte("test"))
		packedBytes, err := packer.Pack(msg)
		assert.Error(t, err)
		assert.Nil(t, packedBytes)
	})

	t.Run("when size is too big", func(t *testing.T) {
		r := bytes.NewBuffer(nil)
		assert.NoError(t, binary.Write(r, binary.BigEndian, uint32(packer.MaxDataSize+1)))
		assert.NoError(t, binary.Write(r, binary.BigEndian, uint32(1)))
		assert.NoError(t, binary.Write(r, binary.BigEndian, []byte("test")))
		msg, err := packer.Unpack(r)
		assert.Error(t, err)
		assert.Nil(t, msg)
	})
}
