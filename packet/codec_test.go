package packet

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStringCodec_Encode(t *testing.T) {
	c := &StringCodec{}
	b, err := c.Encode("hello")
	assert.NoError(t, err)
	assert.Equal(t, []byte("hello"), b)

	b, err = c.Encode(true)
	assert.Error(t, err)
	assert.Nil(t, b)
}

func TestStringCodec_Decode(t *testing.T) {
	c := &StringCodec{}
	data := []byte("hello")
	var v string
	assert.NoError(t, c.Decode(data, &v))
	assert.Equal(t, string(data), v)

	var v2 int
	assert.Error(t, c.Decode(data, &v2))
}

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
