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

func TestRouter_routeLoop(t *testing.T) {
	t.Run("when session is closed", func(t *testing.T) {
		rt := newRouter()

		sess := newSession(nil, &SessionOption{})
		sess.close()
		rt.routeLoop(sess)
	})
	t.Run("when reqQueue is closed", func(t *testing.T) {
		rt := newRouter()

		sess := newSession(nil, &SessionOption{})
		close(sess.reqQueue)
		rt.routeLoop(sess)
	})
	t.Run("when received a nil request", func(t *testing.T) {
		rt := newRouter()

		reqCh := make(chan *message.Entry)
		go func() {
			reqCh <- nil
			close(reqCh)
		}()
		sess := newSession(nil, &SessionOption{})
		go func() {
			sess.reqQueue <- nil
			sess.close()
		}()
		rt.routeLoop(sess) // should not call to handler
	})
	t.Run("when received a non-nil request", func(t *testing.T) {
		t.Run("when handler returns error", func(t *testing.T) {
			rt := newRouter()

			rt.register(1, func(ctx *Context) (*message.Entry, error) {
				assert.EqualValues(t, ctx.Message().ID, 1)
				assert.EqualValues(t, len(ctx.Message().Data), 4)
				assert.Equal(t, ctx.Message().Data, []byte("test"))
				return nil, fmt.Errorf("some err")
			})

			entry := &message.Entry{
				ID:   1,
				Data: []byte("test"),
			}
			sess := newSession(nil, &SessionOption{})
			go func() {
				sess.reqQueue <- entry
				sess.close()
			}()
			rt.routeLoop(sess) // should receive entry only once
		})
		t.Run("when handler returns nil", func(t *testing.T) {
			rt := newRouter()

			rt.register(1, nilHandler)

			entry := &message.Entry{
				ID:   1,
				Data: []byte("test"),
			}
			sess := newSession(nil, &SessionOption{})
			go func() {
				sess.reqQueue <- entry
				sess.close()
			}()
			loopDone := make(chan struct{})
			go func() {
				rt.routeLoop(sess) // should receive entry only once
				close(loopDone)
			}()
			<-loopDone
		})
		t.Run("when handler returns response and send success", func(t *testing.T) {
			rt := newRouter()
			var id = 1

			rt.register(id, func(ctx *Context) (*message.Entry, error) {
				return &message.Entry{}, nil
			})

			sess := newSession(nil, &SessionOption{})
			go func() { <-sess.respQueue }()

			entry := &message.Entry{
				ID:   id,
				Data: []byte("test"),
			}
			loopDone := make(chan struct{})
			go func() { sess.reqQueue <- entry }()
			go func() {
				rt.routeLoop(sess)
				close(loopDone)
			}()
			time.Sleep(time.Millisecond * 5)
			sess.close()
			<-loopDone
		})
		t.Run("when handler returns response but send failed", func(t *testing.T) {
			rt := newRouter()
			var id = 1

			rt.register(id, func(ctx *Context) (*message.Entry, error) {
				defer ctx.Session().close()
				return &message.Entry{}, nil
			})

			sess := newSession(nil, &SessionOption{})
			go func() { <-sess.respQueue }()

			entry := &message.Entry{
				ID:   id,
				Data: []byte("test"),
			}
			loopDone := make(chan struct{})
			go func() { sess.reqQueue <- entry }()
			go func() {
				rt.routeLoop(sess)
				close(loopDone)
			}()
			<-loopDone
		})
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

		msg := &message.Entry{
			ID:   1,
			Data: []byte("test"),
		}
		sess := newSession(nil, &SessionOption{})
		resp, err := rt.handleReq(sess, msg)
		assert.Nil(t, err)
		assert.Nil(t, resp)
	})
	t.Run("when handler and middlewares found", func(t *testing.T) {
		rt := newRouter()
		var id = 1
		rt.register(id, nilHandler, func(next HandlerFunc) HandlerFunc {
			return func(ctx *Context) (*message.Entry, error) { return next(ctx) }
		})

		sess := newSession(nil, &SessionOption{})
		entry := &message.Entry{
			ID:   id,
			Data: []byte("test"),
		}
		resp, err := rt.handleReq(sess, entry)
		assert.Nil(t, err)
		assert.Nil(t, resp)
	})
	t.Run("when handler returns error", func(t *testing.T) {
		rt := newRouter()
		var id = 1
		rt.register(id, func(ctx *Context) (*message.Entry, error) {
			return nil, fmt.Errorf("some err")
		})

		sess := newSession(nil, &SessionOption{})
		msg := &message.Entry{
			ID:   id,
			Data: []byte("test"),
		}

		resp, err := rt.handleReq(sess, msg)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
	// t.Run("when handler returns a non-nil response", func(t *testing.T) {
	// t.Run("when session send resp failed", func(t *testing.T) {
	// 	var id = 1
	// 	rt := newRouter()
	//
	// 	// register route
	// 	rt.register(id, func(ctx *Context) (*message.Entry, error) {
	// 		return &message.Entry{}, nil
	// 	})
	//
	// 	sess := newSession(nil, &SessionOption{})
	// 	sess.Close()
	//
	// 	entry := &message.Entry{
	// 		ID:   id,
	// 		Data: []byte("test"),
	// 	}
	// 	resp, err := rt.handleReq(sess, entry)
	// 	assert.Error(t, err)
	// })
	// t.Run("when session send resp without error", func(t *testing.T) {
	// 	rt := newRouter()
	// 	var id = 1
	//
	// 	rt.register(id, func(ctx *Context) (*message.Entry, error) {
	// 		return &message.Entry{}, nil
	// 	})
	//
	// 	sess := newSession(nil, &SessionOption{})
	// 	go func() { <-sess.respQueue }()
	//
	// 	entry := &message.Entry{
	// 		ID:   id,
	// 		Data: []byte("test"),
	// 	}
	// 	err := rt.handleReq(sess, entry)
	// 	assert.NoError(t, err)
	// })
	// })
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
