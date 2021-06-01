package packet

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDefaultCodec_Encode(t *testing.T) {
	c := &StringCodec{}
	b, err := c.Encode("hello")
	assert.NoError(t, err)
	assert.Equal(t, []byte("hello"), b)
}

func TestDefaultCodec_Decode(t *testing.T) {
	c := &StringCodec{}
	data := []byte("hello")
	var v string
	err := c.Decode(data, &v)
	assert.NoError(t, err)
	assert.Equal(t, string(data), v)
}
