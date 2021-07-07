package easytcp

import (
	"fmt"
	"github.com/DarthPestilane/easytcp/message"
	"github.com/stretchr/testify/assert"
	"reflect"
	"runtime"
	"testing"
)

func TestNewRouter(t *testing.T) {
	rt := NewRouter()
	assert.NotNil(t, rt.globalMiddlewares)
}

func TestRouter_RouteLoop(t *testing.T) {
	t.Run("when session is closed", func(t *testing.T) {
		rt := NewRouter()

		sess := NewSession(nil, &SessionOption{})
		sess.Close()
		rt.RouteLoop(sess)
	})
	t.Run("when received a nil request", func(t *testing.T) {
		rt := NewRouter()

		reqCh := make(chan *message.Entry)
		go func() {
			reqCh <- nil
			close(reqCh)
		}()
		sess := NewSession(nil, &SessionOption{})
		go func() {
			sess.reqQueue <- nil
			sess.Close()
		}()
		rt.RouteLoop(sess) // should not call to handler
	})
	t.Run("when received a non-nil request", func(t *testing.T) {
		t.Run("when handler returns error", func(t *testing.T) {
			rt := NewRouter()

			rt.Register(1, func(ctx *Context) (*message.Entry, error) {
				assert.EqualValues(t, ctx.MsgID(), 1)
				assert.EqualValues(t, ctx.MsgSize(), 4)
				assert.Equal(t, ctx.MsgData(), []byte("test"))
				return nil, fmt.Errorf("some err")
			})

			entry := &message.Entry{
				ID:   1,
				Data: []byte("test"),
			}
			sess := NewSession(nil, &SessionOption{})
			go func() {
				sess.reqQueue <- entry
				sess.Close()
			}()
			rt.RouteLoop(sess) // should receive entry only once
		})
		t.Run("when handler returns no error", func(t *testing.T) {
			rt := NewRouter()

			rt.Register(1, nilHandler)

			entry := &message.Entry{
				ID:   1,
				Data: []byte("test"),
			}
			sess := NewSession(nil, &SessionOption{})
			go func() {
				sess.reqQueue <- entry
				sess.Close()
			}()
			loopDone := make(chan struct{})
			go func() {
				rt.RouteLoop(sess) // should receive entry only once
				close(loopDone)
			}()
			<-loopDone
		})
	})
}

func TestRouter_Register(t *testing.T) {
	rt := NewRouter()

	var id uint = 1

	rt.Register(id, nil)
	_, ok := rt.handlerMapper.Load(id)
	assert.False(t, ok)
	_, ok = rt.middlewaresMapper.Load(id)
	assert.False(t, ok)

	h := nilHandler
	m1 := func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) (*message.Entry, error) {
			return next(ctx)
		}
	}
	m2 := func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) (*message.Entry, error) {
			return next(ctx)
		}
	}
	rt.Register(id, h, m1, nil, m2)
	v, ok := rt.handlerMapper.Load(id)
	assert.True(t, ok)
	expect := runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
	actual := runtime.FuncForPC(reflect.ValueOf(v).Pointer()).Name()
	assert.Equal(t, expect, actual)
	v, ok = rt.middlewaresMapper.Load(id)
	assert.True(t, ok)
	mhs, ok := v.([]MiddlewareFunc)
	assert.True(t, ok)
	expects := []MiddlewareFunc{m1, m2}
	for i, mh := range mhs {
		expect := runtime.FuncForPC(reflect.ValueOf(expects[i]).Pointer()).Name()
		actual := runtime.FuncForPC(reflect.ValueOf(mh).Pointer()).Name()
		assert.Equal(t, expect, actual)
	}
}

func TestRouter_RegisterMiddleware(t *testing.T) {
	rt := NewRouter()

	rt.RegisterMiddleware()
	assert.Len(t, rt.globalMiddlewares, 0)

	rt.RegisterMiddleware(nil, nil)
	assert.Len(t, rt.globalMiddlewares, 0)

	m1 := func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) (*message.Entry, error) {
			return next(ctx)
		}
	}
	m2 := func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) (*message.Entry, error) {
			return next(ctx)
		}
	}
	m3 := func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) (*message.Entry, error) {
			return next(ctx)
		}
	}
	rt.RegisterMiddleware(m1, m2)
	assert.Len(t, rt.globalMiddlewares, 2)

	rt.RegisterMiddleware(m3)
	assert.Len(t, rt.globalMiddlewares, 3)

	expects := []MiddlewareFunc{m1, m2, m3}
	for i, m := range rt.globalMiddlewares {
		expect := runtime.FuncForPC(reflect.ValueOf(expects[i]).Pointer()).Name()
		actual := runtime.FuncForPC(reflect.ValueOf(m).Pointer()).Name()
		assert.Equal(t, expect, actual)
	}
}

func TestRouter_handleReq(t *testing.T) {
	t.Run("when handler and middlewares not found", func(t *testing.T) {
		rt := NewRouter()

		msg := &message.Entry{
			ID:   1,
			Data: []byte("test"),
		}
		sess := NewSession(nil, &SessionOption{})
		assert.Nil(t, rt.handleReq(sess, msg))
	})
	t.Run("when handler and middlewares found", func(t *testing.T) {
		rt := NewRouter()
		var id uint = 1
		rt.Register(id, nilHandler, func(next HandlerFunc) HandlerFunc {
			return func(ctx *Context) (*message.Entry, error) { return next(ctx) }
		})

		sess := NewSession(nil, &SessionOption{})
		entry := &message.Entry{
			ID:   id,
			Data: []byte("test"),
		}

		assert.Nil(t, rt.handleReq(sess, entry))
	})
	t.Run("when handler returns error", func(t *testing.T) {
		rt := NewRouter()
		var id uint = 1
		rt.Register(id, func(ctx *Context) (*message.Entry, error) {
			return nil, fmt.Errorf("some err")
		})

		sess := NewSession(nil, &SessionOption{})
		msg := &message.Entry{
			ID:   id,
			Data: []byte("test"),
		}

		err := rt.handleReq(sess, msg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "handler err")
	})
	t.Run("when handler returns a non-nil response", func(t *testing.T) {
		t.Run("when session send resp failed", func(t *testing.T) {
			var id uint = 1
			rt := NewRouter()

			// register route
			rt.Register(id, func(ctx *Context) (*message.Entry, error) {
				return &message.Entry{}, nil
			})

			sess := NewSession(nil, &SessionOption{})
			close(sess.respQueue)

			entry := &message.Entry{
				ID:   id,
				Data: []byte("test"),
			}
			err := rt.handleReq(sess, entry)
			assert.Error(t, err)
		})
		t.Run("when session send resp without error", func(t *testing.T) {
			rt := NewRouter()
			var id uint = 1

			rt.Register(id, func(ctx *Context) (*message.Entry, error) {
				return &message.Entry{}, nil
			})

			sess := NewSession(nil, &SessionOption{})
			go func() {
				<-sess.respQueue
			}()

			message := &message.Entry{
				ID:   id,
				Data: []byte("test"),
			}
			err := rt.handleReq(sess, message)
			assert.NoError(t, err)
		})
	})
}

func TestRouter_wrapHandlers(t *testing.T) {
	rt := NewRouter()
	t.Run("it works when there's no handler nor middleware", func(t *testing.T) {
		wrap := rt.wrapHandlers(nil, nil)
		resp, err := wrap(nil)
		assert.NoError(t, err)
		assert.Nil(t, resp)
	})
	t.Run("it should invoke handlers in the right order", func(t *testing.T) {
		result := make([]string, 0)

		middles := []MiddlewareFunc{
			func(next HandlerFunc) HandlerFunc {
				return func(ctx *Context) (*message.Entry, error) {
					result = append(result, "m1-before")
					return next(ctx)
				}
			},
			func(next HandlerFunc) HandlerFunc {
				return func(ctx *Context) (*message.Entry, error) {
					result = append(result, "m2-before")
					resp, err := next(ctx)
					result = append(result, "m2-after")
					return resp, err
				}
			},
			func(next HandlerFunc) HandlerFunc {
				return func(ctx *Context) (*message.Entry, error) {
					resp, err := next(ctx)
					result = append(result, "m3-after")
					return resp, err
				}
			},
		}
		var handler HandlerFunc = func(ctx *Context) (*message.Entry, error) {
			result = append(result, "done")
			msg := &message.Entry{
				ID:   2,
				Data: []byte("done"),
			}
			return msg, nil
		}

		wrap := rt.wrapHandlers(handler, middles)
		resp, err := wrap(nil)
		assert.NoError(t, err)
		assert.EqualValues(t, resp.Data, "done")
		assert.Equal(t, result, []string{"m1-before", "m2-before", "done", "m3-after", "m2-after"})
	})
}

func TestRouter_PrintHandlers(t *testing.T) {
	t.Run("when there's no route registered", func(t *testing.T) {
		rt := NewRouter()
		rt.PrintHandlers("localhost")
	})
	t.Run("when there are routes registered", func(t *testing.T) {
		rt := NewRouter()
		rt.Register(1234, nilHandler)
		rt.Register(12345, nilHandler)
		rt.Register(123456, nilHandler)
		rt.Register(12345678, nilHandler)
		rt.PrintHandlers("localhost")
	})
}

func TestRouter_SetNotFoundHandler(t *testing.T) {
	rt := NewRouter()
	assert.Nil(t, rt.notFoundHandler)
	rt.SetNotFoundHandler(func(ctx *Context) (*message.Entry, error) {
		return nil, nil
	})
	assert.NotNil(t, rt.notFoundHandler)
}
