package easytcp

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMessage_GetSetAndRemove(t *testing.T) {
	msg := &Message{}
	msg.Set("key", "test")

	v, ok := msg.Get("key")
	assert.True(t, ok)
	assert.Equal(t, v, "test")

	v, ok = msg.Get("not-found")
	assert.False(t, ok)
	assert.Nil(t, v)

	msg.Remove("key")
	v, ok = msg.Get("key")
	assert.False(t, ok)
	assert.Nil(t, v)
}

func TestMessage_MustGet(t *testing.T) {
	msg := &Message{}
	msg.Set("key", "test")

	v := msg.MustGet("key")
	assert.Equal(t, v, "test")

	assert.Panics(t, func() { msg.MustGet("not-found") })
}
