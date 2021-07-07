package easytcp

import (
	"github.com/stretchr/testify/assert"
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
