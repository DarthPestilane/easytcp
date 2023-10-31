package easytcp

import (
	"context"
	"fmt"
	"github.com/DarthPestilane/easytcp/internal/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func newTestContext(sess *session, reqMsg *Message) *routeContext {
	ctx := newContext()
	ctx.session = sess
	ctx.reqMsg = reqMsg
	return ctx
}

func Test_routeContext_Deadline(t *testing.T) {
	c := newTestContext(nil, nil)
	dl, ok := c.Deadline()
	assert.False(t, ok)
	assert.Zero(t, dl)
}

func Test_routeContext_Done(t *testing.T) {
	c := newTestContext(nil, nil)
	done := c.Done()
	assert.Nil(t, done)
}

func Test_routeContext_Err(t *testing.T) {
	c := newTestContext(nil, nil)
	assert.Nil(t, c.Err())
}

func Test_routeContext_Value(t *testing.T) {
	c := newTestContext(nil, nil)
	assert.Nil(t, c.Value("not found"))
	c.Set("found", true)
	assert.True(t, c.Value("found").(bool))

	assert.Nil(t, c.Value(123))
}

func Test_routeContext_Get(t *testing.T) {
	c := newTestContext(nil, nil)
	v, ok := c.Get("not found")
	assert.False(t, ok)
	assert.Nil(t, v)

	c.Set("found", true)
	v, ok = c.Get("found")
	assert.True(t, ok)
	assert.True(t, v.(bool))
}

func Test_routeContext_Set(t *testing.T) {
	c := newTestContext(nil, nil)
	c.Set("found", true)
	v, ok := c.storage["found"]
	assert.True(t, ok)
	assert.True(t, v.(bool))
}

func Test_routeContext_Remove(t *testing.T) {
	c := newTestContext(nil, nil)
	c.Set("found", true)
	c.Remove("found")
	v, ok := c.Get("found")
	assert.False(t, ok)
	assert.Nil(t, v)
}

func Test_routeContext_Bind(t *testing.T) {
	t.Run("when session has codec", func(t *testing.T) {
		reqMsg := NewMessage(1, []byte(`{"data":"test"}`))
		sess := newSession(nil, &sessionOption{Codec: &JsonCodec{}})

		c := newTestContext(sess, reqMsg)
		data := make(map[string]string)
		assert.NoError(t, c.Bind(&data))
		assert.EqualValues(t, data["data"], "test")

		// when dst is invalid
		var dst string
		assert.Error(t, c.Bind(&dst))
	})
	t.Run("when session hasn't codec", func(t *testing.T) {
		reqMsg := NewMessage(1, []byte("test"))
		sess := newSession(nil, &sessionOption{})

		c := newTestContext(sess, reqMsg)
		var data string
		assert.Error(t, c.Bind(&data))
		assert.Empty(t, data)
	})
}

func Test_routeContext_Session(t *testing.T) {
	sess := newSession(nil, &sessionOption{})

	c := newTestContext(sess, nil)
	assert.Equal(t, c.Session(), sess)
}

func Test_routeContext_SetResponse(t *testing.T) {
	t.Run("when session hasn't codec", func(t *testing.T) {
		reqMsg := NewMessage(1, []byte("test"))
		sess := newSession(nil, &sessionOption{})

		c := newTestContext(sess, reqMsg)
		err := c.SetResponse(1, []string{"invalid", "data"})
		assert.Error(t, err)
		assert.Nil(t, c.respMsg)
	})
	t.Run("when encode failed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		reqMsg := &Message{}
		codec := mock.NewMockCodec(ctrl)
		codec.EXPECT().Encode(gomock.Any()).Return(nil, fmt.Errorf("some err"))
		sess := newSession(nil, &sessionOption{Codec: codec})

		c := newTestContext(sess, reqMsg)
		err := c.SetResponse(1, "test")
		assert.Error(t, err)
		assert.Nil(t, c.respMsg)
	})
	t.Run("when succeed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		reqMsg := NewMessage(1, []byte("test"))
		codec := mock.NewMockCodec(ctrl)
		codec.EXPECT().Encode(gomock.Any()).Return([]byte("test"), nil)
		sess := newSession(nil, &sessionOption{Codec: codec})

		c := newTestContext(sess, reqMsg)
		err := c.SetResponse(1, "test")
		assert.NoError(t, err)
		assert.Equal(t, c.respMsg, reqMsg)
	})
}

func Test_routeContext_Send(t *testing.T) {
	t.Run("when success", func(t *testing.T) {
		sess := newSession(nil, &sessionOption{})
		ctx := newTestContext(sess, nil)
		ctx.SetResponseMessage(NewMessage(1, []byte("test")))
		go ctx.Send()
		ctx2 := <-sess.respQueue
		assert.Equal(t, ctx, ctx2)
	})
}

func Test_routeContext_SendTo(t *testing.T) {
	t.Run("when success", func(t *testing.T) {
		sess1 := newSession(nil, &sessionOption{})
		sess2 := newSession(nil, &sessionOption{})
		ctx := newTestContext(sess1, nil)
		ctx.SetResponseMessage(NewMessage(1, []byte("test")))
		go ctx.SendTo(sess2)
		ctx2 := <-sess2.respQueue
		assert.Equal(t, ctx, ctx2)
	})
}

func Test_routeContext_reset(t *testing.T) {
	sess := newSession(nil, &sessionOption{})
	reqMsg := NewMessage(1, []byte("test"))
	ctx := newTestContext(sess, reqMsg)
	ctx.reset()
	assert.Equal(t, ctx.rawCtx, context.Background())
	assert.Nil(t, ctx.session)
	assert.Nil(t, ctx.reqMsg)
	assert.Nil(t, ctx.respMsg)
	assert.Empty(t, ctx.storage)
}

func Test_routeContext_Copy(t *testing.T) {
	ctx := newTestContext(nil, nil)
	ctx.SetResponseMessage(NewMessage(1, []byte("resp origin")))

	ctx2 := ctx.Copy()
	ctx2.SetResponseMessage(NewMessage(2, []byte("resp copy")))

	assert.EqualValues(t, ctx.respMsg.ID(), 1)
	assert.Equal(t, ctx.respMsg.Data(), []byte("resp origin"))
	assert.EqualValues(t, ctx2.Response().ID(), 2)
	assert.Equal(t, ctx2.Response().Data(), []byte("resp copy"))
}

func Test_routeContext_MustSetResponse(t *testing.T) {
	t.Run("when session hasn't codec", func(t *testing.T) {
		reqMsg := NewMessage(1, []byte("test"))
		sess := newSession(nil, &sessionOption{})

		c := newTestContext(sess, reqMsg)
		assert.Panics(t, func() {
			c.MustSetResponse(1, []string{"invalid", "data"})
		})
	})
	t.Run("when encode failed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		reqMsg := &Message{}
		codec := mock.NewMockCodec(ctrl)
		codec.EXPECT().Encode(gomock.Any()).Return(nil, fmt.Errorf("some err"))
		sess := newSession(nil, &sessionOption{Codec: codec})

		c := newTestContext(sess, reqMsg)
		assert.Panics(t, func() {
			c.MustSetResponse(1, "test")
		})
	})
	t.Run("when succeed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		reqMsg := NewMessage(1, []byte("test"))
		codec := mock.NewMockCodec(ctrl)
		codec.EXPECT().Encode(gomock.Any()).Return([]byte("test"), nil)
		sess := newSession(nil, &sessionOption{Codec: codec})

		c := newTestContext(sess, reqMsg)
		assert.NotPanics(t, func() {
			assert.Equal(t, c.MustSetResponse(1, "test"), c)
		})
	})
}

func Test_routeContext_SetSession(t *testing.T) {
	sess := newSession(nil, &sessionOption{})
	c := newTestContext(nil, nil)
	assert.Equal(t, c.SetSession(sess), c)
	assert.Equal(t, c.Session(), sess)
}

func Test_routeContext_SetRequest(t *testing.T) {
	t.Run("when session hasn't codec", func(t *testing.T) {
		sess := newSession(nil, &sessionOption{})
		c := newTestContext(sess, nil)
		err := c.SetRequest(1, []string{"invalid", "data"})
		assert.Error(t, err)
		assert.Nil(t, c.reqMsg)
	})
	t.Run("when encode failed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		codec := mock.NewMockCodec(ctrl)
		codec.EXPECT().Encode(gomock.Any()).Return(nil, fmt.Errorf("some err"))
		sess := newSession(nil, &sessionOption{Codec: codec})

		c := newTestContext(sess, nil)
		err := c.SetRequest(1, "test")
		assert.Error(t, err)
		assert.Nil(t, c.reqMsg)
	})
	t.Run("when succeed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		reqMsg := NewMessage(1, []byte("test"))
		codec := mock.NewMockCodec(ctrl)
		codec.EXPECT().Encode(gomock.Any()).Return([]byte("test"), nil)
		sess := newSession(nil, &sessionOption{Codec: codec})

		c := newTestContext(sess, nil)
		err := c.SetRequest(1, "test")
		assert.NoError(t, err)
		assert.Equal(t, c.reqMsg, reqMsg)
	})
}

func Test_routeContext_MustSetRequest(t *testing.T) {
	t.Run("when session hasn't codec", func(t *testing.T) {
		sess := newSession(nil, &sessionOption{})

		c := newTestContext(sess, nil)
		assert.Panics(t, func() {
			c.MustSetRequest(1, []string{"invalid", "data"})
		})
	})
	t.Run("when encode failed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		codec := mock.NewMockCodec(ctrl)
		codec.EXPECT().Encode(gomock.Any()).Return(nil, fmt.Errorf("some err"))
		sess := newSession(nil, &sessionOption{Codec: codec})

		c := newTestContext(sess, nil)
		assert.Panics(t, func() {
			c.MustSetRequest(1, "test")
		})
	})
	t.Run("when succeed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		codec := mock.NewMockCodec(ctrl)
		codec.EXPECT().Encode(gomock.Any()).Return([]byte("test"), nil)
		sess := newSession(nil, &sessionOption{Codec: codec})

		c := newTestContext(sess, nil)
		assert.NotPanics(t, func() {
			assert.Equal(t, c.MustSetRequest(1, "test"), c)
		})
	})
}

func Test_routeContext_SetRequestMessage(t *testing.T) {
	reqMsg := NewMessage(1, []byte("test"))
	c := newContext()
	c.SetRequestMessage(reqMsg)
	assert.Equal(t, c.reqMsg, reqMsg)
}
