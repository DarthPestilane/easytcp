package message

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEntry_GetSet(t *testing.T) {
	entry := &Entry{}
	entry.Set("key", "test")

	v, ok := entry.Get("key")
	assert.True(t, ok)
	assert.Equal(t, v, "test")

	v, ok = entry.Get("not-found")
	assert.False(t, ok)
	assert.Nil(t, v)
}

func TestEntry_MustGet(t *testing.T) {
	entry := &Entry{}
	entry.Set("key", "test")

	v := entry.MustGet("key")
	assert.Equal(t, v, "test")

	assert.Panics(t, func() { entry.MustGet("not-found") })
}
