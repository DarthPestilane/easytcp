package easytcp

import (
	"github.com/DarthPestilane/easytcp/internal/test_data/msgpack"
	"github.com/DarthPestilane/easytcp/internal/test_data/pb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"testing"
)

func TestJsonCodec_Decode(t *testing.T) {
	c := &JsonCodec{}
	data := []byte(`{"id": 1}`)
	var v struct {
		Id int `json:"id"`
	}
	assert.NoError(t, c.Decode(data, &v))
	assert.EqualValues(t, v.Id, 1)
}

func TestJsonCodec_Encode(t *testing.T) {
	c := &JsonCodec{}
	v := struct {
		Id int `json:"id"`
	}{Id: 1}
	b, err := c.Encode(v)
	assert.NoError(t, err)
	assert.JSONEq(t, string(b), `{"id": 1}`)
}

func TestProtobufCodec(t *testing.T) {
	c := &ProtobufCodec{}
	t.Run("when encode/decode with invalid params", func(t *testing.T) {
		// encoding
		b, err := c.Encode(123)
		assert.Error(t, err)
		assert.Nil(t, b)

		// decoding
		var v int
		assert.Error(t, c.Decode([]byte("test"), &v))
	})
	t.Run("when succeed", func(t *testing.T) {
		// encoding
		v := &pb.Sample{Foo: "bar", Bar: 33}
		b, err := c.Encode(v)
		assert.NoError(t, err)
		assert.NotNil(t, b)

		// decoding
		sample := &pb.Sample{}
		assert.NoError(t, c.Decode(b, sample))

		// comparing
		assert.True(t, proto.Equal(v, sample))
	})
}

func TestMsgpackCodec(t *testing.T) {
	item := msgpack.Sample{
		Foo: "test",
		Bar: 1,
		Baz: map[int]string{2: "2"},
	}
	c := &MsgpackCodec{}
	b, err := c.Encode(item)
	assert.NoError(t, err)
	assert.NotNil(t, b)

	var itemDec msgpack.Sample
	assert.NoError(t, c.Decode(b, &itemDec))
	assert.Equal(t, item, itemDec)
}
