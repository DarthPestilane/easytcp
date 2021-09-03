package easytcp

import (
	"fmt"
	"github.com/DarthPestilane/easytcp/message"
	"github.com/stretchr/testify/assert"
	"reflect"
	"runtime"
	"testing"
	"time"
)

func TestRouter_consumeRequest(t *testing.T) {
	t.Run("when router is closed", func(t *testing.T) {
		rt := newRouter(100)

		done := make(chan struct{})
		go func() {
			rt.consumeRequest()
			close(done)
		}()
		time.Sleep(time.Millisecond * 5)
		go rt.stop()
		<-done
		_, ok := <-rt.reqCtxQueue
		assert.False(t, ok)
	})
	t.Run("when router reqCtxQueue is closed", func(t *testing.T) {
		rt := newRouter(100)

		done := make(chan struct{})
		go func() {
			rt.consumeRequest()
			close(done)
		}()
		time.Sleep(time.Millisecond * 5)
		close(rt.reqCtxQueue)
		<-done
	})
	t.Run("when ctx session is closed", func(t *testing.T) {
		rt := newRouter(100)
		done := make(chan struct{})
		go func() {
			rt.consumeRequest()
			close(done)
		}()
		time.Sleep(time.Millisecond * 5)
		sess := newSession(nil, &SessionOption{})
		sess.close()
		rt.reqCtxQueue <- &Context{session: sess}
		time.Sleep(time.Millisecond * 5)
		rt.stop()
		<-done
	})
	t.Run("when ctx message entry is nil", func(t *testing.T) {
		rt := newRouter(100)
		done := make(chan struct{})
		go func() {
			rt.consumeRequest()
			close(done)
		}()
		time.Sleep(time.Millisecond * 5)
		sess := newSession(nil, &SessionOption{})
		rt.reqCtxQueue <- &Context{session: sess}
		time.Sleep(time.Millisecond * 5)
		rt.stop()
		<-done
	})
	t.Run("when handler returns error", func(t *testing.T) {
		rt := newRouter(100)
		done := make(chan struct{})
		go func() {
			rt.consumeRequest()
			close(done)
		}()
		time.Sleep(time.Millisecond * 5)

		rt.register(1, func(ctx *Context) (*message.Entry, error) {
			return nil, fmt.Errorf("some err")
		})

		sess := newSession(nil, &SessionOption{})
		entry := &message.Entry{ID: 1, Data: []byte("test")}
		rt.reqCtxQueue <- &Context{session: sess, reqMsg: entry}
		time.Sleep(time.Millisecond * 5)
		rt.stop()
		<-done
	})
	t.Run("when handler returns nil response", func(t *testing.T) {
		rt := newRouter(100)
		done := make(chan struct{})
		go func() {
			rt.consumeRequest()
			close(done)
		}()
		time.Sleep(time.Millisecond * 5)

		rt.register(1, func(ctx *Context) (*message.Entry, error) {
			return nil, nil
		})

		sess := newSession(nil, &SessionOption{})
		entry := &message.Entry{ID: 1, Data: []byte("test")}
		rt.reqCtxQueue <- &Context{session: sess, reqMsg: entry}
		time.Sleep(time.Millisecond * 5)
		rt.stop()
		<-done
	})
	t.Run("when send response failed", func(t *testing.T) {
		rt := newRouter(100)
		done := make(chan struct{})
		go func() {
			rt.consumeRequest()
			close(done)
		}()
		time.Sleep(time.Millisecond * 5)

		rt.register(1, func(ctx *Context) (*message.Entry, error) {
			defer ctx.session.close()
			return &message.Entry{}, nil
		})

		sess := newSession(nil, &SessionOption{})
		entry := &message.Entry{ID: 1, Data: []byte("test")}
		rt.reqCtxQueue <- &Context{session: sess, reqMsg: entry}
		time.Sleep(time.Millisecond * 5)
		rt.stop()
		<-done
	})
}

func TestRouter_register(t *testing.T) {
	rt := newRouter()

	var id = 1

	rt.register(id, nil)
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
	rt.register(id, h, m1, nil, m2)
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

func TestRouter_registerMiddleware(t *testing.T) {
	rt := newRouter()

	rt.registerMiddleware()
	assert.Len(t, rt.globalMiddlewares, 0)

	rt.registerMiddleware(nil, nil)
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
	rt.registerMiddleware(m1, m2)
	assert.Len(t, rt.globalMiddlewares, 2)

	rt.registerMiddleware(m3)
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
		rt := newRouter()

		entry := &message.Entry{
			ID:   1,
			Data: []byte("test"),
		}
		resp, err := rt.handleRequest(&Context{reqMsg: entry})
		assert.Nil(t, err)
		assert.Nil(t, resp)
	})
	t.Run("when handler and middlewares found", func(t *testing.T) {
		rt := newRouter()
		var id = 1
		rt.register(id, nilHandler, func(next HandlerFunc) HandlerFunc {
			return func(ctx *Context) (*message.Entry, error) { return next(ctx) }
		})

		entry := &message.Entry{
			ID:   id,
			Data: []byte("test"),
		}
		resp, err := rt.handleRequest(&Context{reqMsg: entry})
		assert.Nil(t, err)
		assert.Nil(t, resp)
	})
	t.Run("when handler returns error", func(t *testing.T) {
		rt := newRouter()
		var id = 1
		rt.register(id, func(ctx *Context) (*message.Entry, error) {
			return nil, fmt.Errorf("some err")
		})

		entry := &message.Entry{
			ID:   id,
			Data: []byte("test"),
		}
		resp, err := rt.handleRequest(&Context{reqMsg: entry})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestRouter_wrapHandlers(t *testing.T) {
	rt := newRouter()
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

func TestRouter_printHandlers(t *testing.T) {
	t.Run("when there's no route registered", func(t *testing.T) {
		rt := newRouter()
		rt.printHandlers("localhost")
	})
	t.Run("when there are routes registered", func(t *testing.T) {
		rt := newRouter()
		rt.register(1234, nilHandler)
		rt.register(12345, nilHandler)
		rt.register(123456, nilHandler)
		rt.register(12345678, nilHandler)
		rt.printHandlers("localhost")
	})
}

func TestRouter_setNotFoundHandler(t *testing.T) {
	rt := newRouter()
	assert.Nil(t, rt.notFoundHandler)
	rt.setNotFoundHandler(func(ctx *Context) (*message.Entry, error) {
		return nil, nil
	})
	assert.NotNil(t, rt.notFoundHandler)
}
