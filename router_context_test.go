package easytcp

import (
	"fmt"
	"github.com/DarthPestilane/easytcp/message"
	"github.com/DarthPestilane/easytcp/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func newContext(sess *Session, msg *message.Entry) *Context {
	return &Context{session: sess, reqMsg: msg}
}

func TestContext_Deadline(t *testing.T) {
	c := newContext(nil, nil)
	dl, ok := c.Deadline()
	assert.False(t, ok)
	assert.Zero(t, dl)
}

func TestContext_Done(t *testing.T) {
	c := newContext(nil, nil)
	done := c.Done()
	assert.Nil(t, done)
}

func TestContext_Err(t *testing.T) {
	c := newContext(nil, nil)
	assert.Nil(t, c.Err())
}

func TestContext_Value(t *testing.T) {
	c := newContext(nil, nil)
	assert.Nil(t, c.Value("not found"))

	c.storage.Store("found", true)
	assert.True(t, c.Value("found").(bool))

	assert.Nil(t, c.Value(123))
}

func TestContext_Get(t *testing.T) {
	c := newContext(nil, nil)
	v, ok := c.Get("not found")
	assert.False(t, ok)
	assert.Nil(t, v)

	c.storage.Store("found", true)
	v, ok = c.Get("found")
	assert.True(t, ok)
	assert.True(t, v.(bool))
}

func TestContext_Set(t *testing.T) {
	c := newContext(nil, nil)
	c.Set("found", true)
	v, ok := c.storage.Load("found")
	assert.True(t, ok)
	assert.True(t, v.(bool))
}

func TestContext_Bind(t *testing.T) {
	t.Run("when session has codec", func(t *testing.T) {
		entry := &message.Entry{
			ID:   1,
			Data: []byte(`{"data":"test"}`),
		}
		sess := newSession(nil, &SessionOption{Codec: &JsonCodec{}})

		c := newContext(sess, entry)
		data := make(map[string]string)
		assert.NoError(t, c.Bind(&data))
		assert.EqualValues(t, data["data"], "test")
	})
	t.Run("when session hasn't codec", func(t *testing.T) {
		entry := &message.Entry{
			ID:   1,
			Data: []byte("test"),
		}
		sess := newSession(nil, &SessionOption{})

		c := newContext(sess, entry)
		var data string
		assert.Error(t, c.Bind(&data))
		assert.Empty(t, data)
	})
}

func TestContext_Session(t *testing.T) {
	sess := newSession(nil, &SessionOption{})

	c := newContext(sess, nil)
	assert.Equal(t, c.Session(), sess)
}

type DataStringer struct{}

func (*DataStringer) String() string {
	return "data"
}

func TestContext_Response(t *testing.T) {
	t.Run("when session hasn't codec", func(t *testing.T) {
		t.Run("when response data is invalid", func(t *testing.T) {
			entry := &message.Entry{
				ID:   1,
				Data: []byte("test"),
			}
			sess := newSession(nil, &SessionOption{})

			c := newContext(sess, entry)
			respMsg, err := c.Response(1, []string{"invalid", "data"})
			assert.Error(t, err)
			assert.Nil(t, respMsg)
		})
		t.Run("when response data is a string", func(t *testing.T) {
			entry := &message.Entry{}
			sess := newSession(nil, &SessionOption{})

			c := newContext(sess, entry)
			respMsg, err := c.Response(1, "data")
			assert.NoError(t, err)
			assert.Equal(t, respMsg.Data, []byte("data"))
			assert.EqualValues(t, respMsg.ID, 1)

			data := "ptr"
			respMsg, err = c.Response(1, &data)
			assert.NoError(t, err)
			assert.Equal(t, respMsg.Data, []byte("ptr"))
			assert.EqualValues(t, respMsg.ID, 1)
		})
		t.Run("when response data is []byte", func(t *testing.T) {
			entry := &message.Entry{}
			sess := newSession(nil, &SessionOption{})

			c := newContext(sess, entry)
			respMsg, err := c.Response(1, []byte("data"))
			assert.NoError(t, err)
			assert.Equal(t, respMsg.Data, []byte("data"))
			assert.EqualValues(t, respMsg.ID, 1)

			data := []byte("data")
			respMsg, err = c.Response(1, &data)
			assert.NoError(t, err)
			assert.Equal(t, respMsg.Data, []byte("data"))
			assert.EqualValues(t, respMsg.ID, 1)
		})
		t.Run("when response data is a Stringer", func(t *testing.T) {
			entry := &message.Entry{}
			sess := newSession(nil, &SessionOption{})

			data := &DataStringer{}
			c := newContext(sess, entry)
			respMsg, err := c.Response(1, data)
			assert.NoError(t, err)
			assert.Equal(t, respMsg.Data, []byte(data.String()))
			assert.EqualValues(t, respMsg.ID, 1)
		})
	})
	t.Run("when encode failed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		entry := &message.Entry{}
		codec := mock.NewMockCodec(ctrl)
		codec.EXPECT().Encode(gomock.Any()).Return(nil, fmt.Errorf("some err"))
		sess := newSession(nil, &SessionOption{Codec: codec})

		c := newContext(sess, entry)
		respMsg, err := c.Response(1, "test")
		assert.Error(t, err)
		assert.Nil(t, respMsg)
	})
	t.Run("when succeed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		entry := &message.Entry{
			ID:   1,
			Data: []byte("test"),
		}
		codec := mock.NewMockCodec(ctrl)
		codec.EXPECT().Encode(gomock.Any()).Return([]byte("test"), nil)
		sess := newSession(nil, &SessionOption{Codec: codec})

		c := newContext(sess, entry)
		respMsg, err := c.Response(1, "test")
		assert.NoError(t, err)
		assert.Equal(t, respMsg, entry)
	})
}
