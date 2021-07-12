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
}
