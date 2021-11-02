package easytcp

import (
	"fmt"
	"github.com/DarthPestilane/easytcp/internal/mock"
	"github.com/DarthPestilane/easytcp/message"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func newContext(sess *session, msg *message.Entry) *routeContext {
	return &routeContext{session: sess, reqEntry: msg}
}

func TestRouteContext_Deadline(t *testing.T) {
	c := newContext(nil, nil)
	dl, ok := c.Deadline()
	assert.False(t, ok)
	assert.Zero(t, dl)
}

func TestRouteContext_Done(t *testing.T) {
	c := newContext(nil, nil)
	done := c.Done()
	assert.Nil(t, done)
}

func TestRouteContext_Err(t *testing.T) {
	c := newContext(nil, nil)
	assert.Nil(t, c.Err())
}

func TestRouteContext_Value(t *testing.T) {
	c := newContext(nil, nil)
	assert.Nil(t, c.Value("not found"))
	c.Set("found", true)
	assert.True(t, c.Value("found").(bool))

	assert.Nil(t, c.Value(123))
}

func TestRouteContext_Get(t *testing.T) {
	c := newContext(nil, nil)
	v, ok := c.Get("not found")
	assert.False(t, ok)
	assert.Nil(t, v)

	c.Set("found", true)
	v, ok = c.Get("found")
	assert.True(t, ok)
	assert.True(t, v.(bool))
}

func TestRouteContext_Set(t *testing.T) {
	c := newContext(nil, nil)
	c.Set("found", true)
	v, ok := c.storage["found"]
	assert.True(t, ok)
	assert.True(t, v.(bool))
}

func TestRouteContext_Remove(t *testing.T) {
	c := newContext(nil, nil)
	c.Set("found", true)
	c.Remove("found")
	v, ok := c.Get("found")
	assert.False(t, ok)
	assert.Nil(t, v)
}

func TestRouteContext_Bind(t *testing.T) {
	t.Run("when session has codec", func(t *testing.T) {
		entry := &message.Entry{
			ID:   1,
			Data: []byte(`{"data":"test"}`),
		}
		sess := newSession(nil, &sessionOption{Codec: &JsonCodec{}})

		c := newContext(sess, entry)
		data := make(map[string]string)
		assert.NoError(t, c.Bind(&data))
		assert.EqualValues(t, data["data"], "test")

		// when dst is invalid
		var dst string
		assert.Error(t, c.Bind(&dst))
	})
	t.Run("when session hasn't codec", func(t *testing.T) {
		entry := &message.Entry{
			ID:   1,
			Data: []byte("test"),
		}
		sess := newSession(nil, &sessionOption{})

		c := newContext(sess, entry)
		var data string
		assert.Error(t, c.Bind(&data))
		assert.Empty(t, data)
	})
}

func TestRouteContext_Session(t *testing.T) {
	sess := newSession(nil, &sessionOption{})

	c := newContext(sess, nil)
	assert.Equal(t, c.Session(), sess)
}

func TestRouteContext_SetResponse(t *testing.T) {
	t.Run("when session hasn't codec", func(t *testing.T) {
		entry := &message.Entry{
			ID:   1,
			Data: []byte("test"),
		}
		sess := newSession(nil, &sessionOption{})

		c := newContext(sess, entry)
		err := c.SetResponse(1, []string{"invalid", "data"})
		assert.Error(t, err)
		assert.Nil(t, c.respEntry)
	})
	t.Run("when encode failed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		entry := &message.Entry{}
		codec := mock.NewMockCodec(ctrl)
		codec.EXPECT().Encode(gomock.Any()).Return(nil, fmt.Errorf("some err"))
		sess := newSession(nil, &sessionOption{Codec: codec})

		c := newContext(sess, entry)
		err := c.SetResponse(1, "test")
		assert.Error(t, err)
		assert.Nil(t, c.respEntry)
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
		sess := newSession(nil, &sessionOption{Codec: codec})

		c := newContext(sess, entry)
		err := c.SetResponse(1, "test")
		assert.NoError(t, err)
		assert.Equal(t, c.respEntry, entry)
	})
}

func TestRouteContext_Send(t *testing.T) {
	t.Run("when success", func(t *testing.T) {
		sess := newSession(nil, &sessionOption{})
		ctx := newContext(sess, nil)
		ctx.SetResponseMessage(&message.Entry{ID: 1, Data: []byte("test")})
		go ctx.Send()
		ctx2 := <-sess.respQueue
		assert.Equal(t, ctx, ctx2)
	})
	t.Run("when response message is nil", func(t *testing.T) {
		sess := newSession(nil, &sessionOption{})
		ctx := newContext(sess, nil)
		ctx.Send()
	})
}

func TestRouteContext_SendTo(t *testing.T) {
	t.Run("when success", func(t *testing.T) {
		sess1 := newSession(nil, &sessionOption{})
		sess2 := newSession(nil, &sessionOption{})
		ctx := newContext(sess1, nil)
		ctx.SetResponseMessage(&message.Entry{ID: 1, Data: []byte("test")})
		go ctx.SendTo(sess2)
		ctx2 := <-sess2.respQueue
		assert.Equal(t, ctx, ctx2)
	})
	t.Run("when response message is nil", func(t *testing.T) {
		sess1 := newSession(nil, &sessionOption{})
		sess2 := newSession(nil, &sessionOption{})
		ctx := newContext(sess1, nil)
		ctx.SendTo(sess2)
	})
}

func TestRouteContext_reset(t *testing.T) {
	ctx := newContext(nil, nil)
	sess := newSession(nil, &sessionOption{})
	entry := &message.Entry{
		ID:   1,
		Data: []byte("test"),
	}
	ctx.reset(sess, entry)
	assert.Equal(t, ctx.session, sess)
	assert.Equal(t, ctx.reqEntry, entry)
	assert.Nil(t, ctx.storage)
	assert.Nil(t, ctx.respEntry)
}

func TestRouteContext_Copy(t *testing.T) {
	ctx := newContext(nil, nil)
	ctx.SetResponseMessage(&message.Entry{ID: 1, Data: []byte("resp origin")})

	ctx2 := ctx.Copy()
	ctx2.SetResponseMessage(&message.Entry{ID: 2, Data: []byte("resp copy")})

	assert.EqualValues(t, ctx.respEntry.ID, 1)
	assert.Equal(t, ctx.respEntry.Data, []byte("resp origin"))
	assert.EqualValues(t, ctx2.Response().ID, 2)
	assert.Equal(t, ctx2.Response().Data, []byte("resp copy"))
}

func Test_routeContext_MustSetResponse(t *testing.T) {
	t.Run("when session hasn't codec", func(t *testing.T) {
		entry := &message.Entry{
			ID:   1,
			Data: []byte("test"),
		}
		sess := newSession(nil, &sessionOption{})

		c := newContext(sess, entry)
		assert.Panics(t, func() {
			c.MustSetResponse(1, []string{"invalid", "data"})
		})
	})
	t.Run("when encode failed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		entry := &message.Entry{}
		codec := mock.NewMockCodec(ctrl)
		codec.EXPECT().Encode(gomock.Any()).Return(nil, fmt.Errorf("some err"))
		sess := newSession(nil, &sessionOption{Codec: codec})

		c := newContext(sess, entry)
		assert.Panics(t, func() {
			c.MustSetResponse(1, "test")
		})
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
		sess := newSession(nil, &sessionOption{Codec: codec})

		c := newContext(sess, entry)
		assert.NotPanics(t, func() {
			c.MustSetResponse(1, "test")
		})
	})
}
