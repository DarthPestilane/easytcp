package packet

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDefaultMsg_Duplicate(t *testing.T) {
	msg := &DefaultMsg{
		ID:   10,
		Size: 10,
		Data: []byte("test"),
	}
	dup := msg.Duplicate()
	assert.Empty(t, dup)
}

func TestDefaultMsg_Setup(t *testing.T) {
	msg := &DefaultMsg{}
	msg.Setup(1, []byte("test"))
	assert.EqualValues(t, msg.GetID(), 1)
	assert.EqualValues(t, msg.GetSize(), len("test"))
	assert.Equal(t, msg.GetData(), []byte("test"))
}
