package easytcp

import (
	"fmt"
	"github.com/DarthPestilane/easytcp/internal/mock"
	"github.com/DarthPestilane/easytcp/message"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func newContext(sess *Session, msg *message.Entry) *Context {
	return &Context{session: sess, reqEntry: msg}
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
	c.Set("found", true)
	assert.True(t, c.Value("found").(bool))

	assert.Nil(t, c.Value(123))
}

func TestContext_Get(t *testing.T) {
	c := newContext(nil, nil)
	v, ok := c.Get("not found")
	assert.False(t, ok)
	assert.Nil(t, v)

	c.Set("found", true)
	v, ok = c.Get("found")
	assert.True(t, ok)
	assert.True(t, v.(bool))
}

func TestContext_MustGet(t *testing.T) {
	c := newContext(nil, nil)
	assert.Panics(t, func() { c.MustGet("not found") })

	c.Set("found", true)
	v := c.MustGet("found")
	assert.True(t, v.(bool))
}

func TestContext_Set(t *testing.T) {
	c := newContext(nil, nil)
	c.Set("found", true)
	v, ok := c.storage["found"]
	assert.True(t, ok)
	assert.True(t, v.(bool))
}

func TestContext_Remove(t *testing.T) {
	c := newContext(nil, nil)
	c.Set("found", true)
	c.Remove("found")
	v, ok := c.Get("found")
	assert.False(t, ok)
	assert.Nil(t, v)
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

		// when dst is invalid
		var dst string
		assert.Error(t, c.Bind(&dst))
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

func TestContext_MustBind(t *testing.T) {
	t.Run("when session has codec", func(t *testing.T) {
		entry := &message.Entry{
			ID:   1,
			Data: []byte(`{"data":"test"}`),
		}
		sess := newSession(nil, &SessionOption{Codec: &JsonCodec{}})
		c := newContext(sess, entry)
		data := make(map[string]string)
		c.MustBind(&data)
		assert.EqualValues(t, data["data"], "test")

		// when dst is invalid
		var dst string
		assert.Panics(t, func() { c.MustBind(&dst) })
	})
	t.Run("when codec is nil", func(t *testing.T) {
		sess := newSession(nil, &SessionOption{})
		c := newContext(sess, &message.Entry{})
		var dst interface{}
		assert.Panics(t, func() { c.MustBind(&dst) }) // should panic
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
			err := c.Response(1, []string{"invalid", "data"})
			assert.Error(t, err)
			assert.Nil(t, c.respEntry)
		})
		t.Run("when response data is a string", func(t *testing.T) {
			entry := &message.Entry{}
			sess := newSession(nil, &SessionOption{})

			c := newContext(sess, entry)
			err := c.Response(1, "data")
			assert.NoError(t, err)
			assert.Equal(t, c.respEntry.Data, []byte("data"))
			assert.EqualValues(t, c.respEntry.ID, 1)

			data := "ptr"
			err = c.Response(1, &data)
			assert.NoError(t, err)
			assert.Equal(t, c.respEntry.Data, []byte("ptr"))
			assert.EqualValues(t, c.respEntry.ID, 1)
		})
		t.Run("when response data is []byte", func(t *testing.T) {
			entry := &message.Entry{}
			sess := newSession(nil, &SessionOption{})

			c := newContext(sess, entry)
			err := c.Response(1, []byte("data"))
			assert.NoError(t, err)
			assert.Equal(t, c.respEntry.Data, []byte("data"))
			assert.EqualValues(t, c.respEntry.ID, 1)

			data := []byte("data")
			err = c.Response(1, &data)
			assert.NoError(t, err)
			assert.Equal(t, c.respEntry.Data, []byte("data"))
			assert.EqualValues(t, c.respEntry.ID, 1)
		})
		t.Run("when response data is a Stringer", func(t *testing.T) {
			entry := &message.Entry{}
			sess := newSession(nil, &SessionOption{})

			data := &DataStringer{}
			c := newContext(sess, entry)
			err := c.Response(1, data)
			assert.NoError(t, err)
			assert.Equal(t, c.respEntry.Data, []byte(data.String()))
			assert.EqualValues(t, c.respEntry.ID, 1)
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
		err := c.Response(1, "test")
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
		sess := newSession(nil, &SessionOption{Codec: codec})

		c := newContext(sess, entry)
		err := c.Response(1, "test")
		assert.NoError(t, err)
		assert.Equal(t, c.respEntry, entry)
	})
}

func TestContext_DecodeTo(t *testing.T) {
	sess := newSession(nil, &SessionOption{Codec: &JsonCodec{}})
	ctx := newContext(sess, &message.Entry{})
	var dst struct {
		Data string `json:"data"`
	}
	assert.NoError(t, ctx.DecodeTo([]byte(`{"data":"test"}`), &dst))
	assert.Equal(t, dst.Data, "test")
}

func TestContext_DecodeTo_when_codec_is_nil(t *testing.T) {
	sess := newSession(nil, &SessionOption{Codec: nil})
	ctx := newContext(sess, &message.Entry{})
	var dst struct {
		Data string `json:"data"`
	}
	assert.Error(t, ctx.DecodeTo([]byte(`{"data":"test"}`), &dst))
	assert.Zero(t, dst.Data)
}

func TestContext_MustDecodeTo(t *testing.T) {
	sess := newSession(nil, &SessionOption{Codec: &JsonCodec{}})
	ctx := newContext(sess, &message.Entry{})
	var dst struct {
		Data string `json:"data"`
	}
	ctx.MustDecodeTo([]byte(`{"data":"test"}`), &dst)
	assert.Equal(t, dst.Data, "test")
}

func TestContext_MustDecodeTo_when_decode_fail(t *testing.T) {
	sess := newSession(nil, &SessionOption{Codec: &JsonCodec{}})
	ctx := newContext(sess, &message.Entry{})
	var dst string
	assert.Panics(t, func() {
		ctx.MustDecodeTo([]byte(`{"data":"test"}`), &dst)
	})
}

func TestContext_SendTo(t *testing.T) {
	sess := newSession(nil, &SessionOption{})
	ctx := newContext(sess, nil)
	sess2 := newSession(nil, &SessionOption{})
	go func() { <-sess2.respQueue }()
	assert.NoError(t, ctx.SendTo(sess2, 1, []byte("test")))
}

func TestContext_SendTo_when_error(t *testing.T) {
	sess := newSession(nil, &SessionOption{})
	ctx := newContext(sess, nil)
	sess2 := newSession(nil, &SessionOption{})
	assert.Error(t, ctx.SendTo(sess2, 1, 1234))
}

func TestContext_Send(t *testing.T) {
	sess := newSession(nil, &SessionOption{})
	ctx := newContext(sess, nil)
	go func() { <-sess.respQueue }()
	assert.NoError(t, ctx.Send(1, []byte("test")))
}

func TestContext_Send_when_error(t *testing.T) {
	sess := newSession(nil, &SessionOption{})
	ctx := newContext(sess, nil)
	assert.Error(t, ctx.Send(1, 1234))
}

func TestContext_SetResponse(t *testing.T) {
	ctx := newContext(nil, nil)
	entry := &message.Entry{
		ID:   1,
		Data: []byte("test"),
	}
	ctx.SetResponse(entry.ID, entry.Data)
	assert.Equal(t, ctx.respEntry, entry)
}

func TestContext_GetResponse(t *testing.T) {
	ctx := newContext(nil, nil)
	entry := &message.Entry{
		ID:   1,
		Data: []byte("test"),
	}
	ctx.SetResponse(entry.ID, entry.Data)
	respEntry := ctx.GetResponse()
	assert.Equal(t, respEntry, entry)
}

func TestContext_reset(t *testing.T) {
	ctx := newContext(nil, nil)
	sess := newSession(nil, &SessionOption{})
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

func TestContext_Copy(t *testing.T) {
	ctx := newContext(nil, nil)
	ctx.SetResponse(1, []byte("resp origin"))

	ctx2 := ctx.Copy()
	ctx2.SetResponse(2, []byte("resp copy"))

	assert.EqualValues(t, ctx.respEntry.ID, 1)
	assert.Equal(t, ctx.respEntry.Data, []byte("resp origin"))
	assert.EqualValues(t, ctx2.respEntry.ID, 2)
	assert.Equal(t, ctx2.respEntry.Data, []byte("resp copy"))
}
