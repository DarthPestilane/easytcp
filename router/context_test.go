package router

import (
	"fmt"
	"github.com/DarthPestilane/easytcp/packet"
	mockPacker "github.com/DarthPestilane/easytcp/packet/mock"
	"github.com/DarthPestilane/easytcp/session/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

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
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		message := &packet.MessageEntry{
			ID:   1,
			Data: []byte(`{"data":"test"}`),
		}
		sess := mock.NewMockSession(ctrl)
		sess.EXPECT().Codec().Return(&packet.JsonCodec{})

		c := newContext(sess, message)
		data := make(map[string]string)
		assert.NoError(t, c.Bind(&data))
		assert.EqualValues(t, data["data"], "test")
	})
	t.Run("when session hasn't codec", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		message := &packet.MessageEntry{
			ID:   1,
			Data: []byte("test"),
		}
		sess := mock.NewMockSession(ctrl)
		sess.EXPECT().Codec().Return(nil)

		c := newContext(sess, message)
		var data string
		assert.Error(t, c.Bind(&data))
		assert.Empty(t, data)
	})
}

func TestContext_SessionID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sess := mock.NewMockSession(ctrl)
	sess.EXPECT().ID().Return("01")

	c := newContext(sess, nil)
	assert.Equal(t, c.SessionID(), "01")
}

type DataStringer struct{}

func (*DataStringer) String() string {
	return "data"
}

func TestContext_Response(t *testing.T) {
	t.Run("when session hasn't codec", func(t *testing.T) {
		t.Run("when response data is invalid", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			message := &packet.MessageEntry{
				ID:   1,
				Data: []byte("test"),
			}
			sess := mock.NewMockSession(ctrl)
			sess.EXPECT().Codec().MinTimes(1).Return(nil)

			c := newContext(sess, message)
			respMsg, err := c.Response(1, []string{"invalid", "data"})
			assert.Error(t, err)
			assert.Nil(t, respMsg)
		})
		t.Run("when response data is a string", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			message := &packet.MessageEntry{}
			sess := mock.NewMockSession(ctrl)
			sess.EXPECT().Codec().MinTimes(1).Return(nil)

			c := newContext(sess, message)
			respMsg, err := c.Response(1, "data")
			assert.NoError(t, err)
			assert.Equal(t, respMsg.Data, []byte("data"))
			assert.EqualValues(t, respMsg.ID, 1)
		})
		t.Run("when response data is []byte", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			message := &packet.MessageEntry{}
			sess := mock.NewMockSession(ctrl)
			sess.EXPECT().Codec().MinTimes(1).Return(nil)

			c := newContext(sess, message)
			respMsg, err := c.Response(1, []byte("data"))
			assert.NoError(t, err)
			assert.Equal(t, respMsg.Data, []byte("data"))
			assert.EqualValues(t, respMsg.ID, 1)
		})
		t.Run("when response data is a Stringer", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			message := &packet.MessageEntry{}
			sess := mock.NewMockSession(ctrl)
			sess.EXPECT().Codec().MinTimes(1).Return(nil)

			data := &DataStringer{}
			c := newContext(sess, message)
			respMsg, err := c.Response(1, data)
			assert.NoError(t, err)
			assert.Equal(t, respMsg.Data, []byte(data.String()))
			assert.EqualValues(t, respMsg.ID, 1)
		})
	})
	t.Run("when encode failed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		message := &packet.MessageEntry{}
		codec := mockPacker.NewMockCodec(ctrl)
		codec.EXPECT().Encode(gomock.Any()).Return(nil, fmt.Errorf("some err"))

		sess := mock.NewMockSession(ctrl)
		sess.EXPECT().Codec().MinTimes(1).Return(codec)

		c := newContext(sess, message)
		respMsg, err := c.Response(1, "test")
		assert.Error(t, err)
		assert.Nil(t, respMsg)
	})
	t.Run("when succeed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		message := &packet.MessageEntry{
			ID:   1,
			Data: []byte("test"),
		}
		codec := mockPacker.NewMockCodec(ctrl)
		codec.EXPECT().Encode(gomock.Any()).Return([]byte("test"), nil)

		sess := mock.NewMockSession(ctrl)
		sess.EXPECT().Codec().MinTimes(1).Return(codec)

		c := newContext(sess, message)
		respMsg, err := c.Response(1, "test")
		assert.NoError(t, err)
		assert.Equal(t, respMsg, message)
	})
}
