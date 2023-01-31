package easytcp

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"runtime"
	"testing"
)

func TestRouter_register(t *testing.T) {
	rt := newRouter()

	var id = 1

	rt.register(id, nil)
	_, ok := rt.handlerMapper[id]
	assert.False(t, ok)
	_, ok = rt.middlewaresMapper[id]
	assert.False(t, ok)

	h := nilHandler
	m1 := func(next HandlerFunc) HandlerFunc {
		return func(ctx Context) {
			next(ctx)
		}
	}
	m2 := func(next HandlerFunc) HandlerFunc {
		return func(ctx Context) {
			next(ctx)
		}
	}
	rt.register(id, h, m1, nil, m2)
	v, ok := rt.handlerMapper[id]
	assert.True(t, ok)
	expect := runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
	actual := runtime.FuncForPC(reflect.ValueOf(v).Pointer()).Name()
	assert.Equal(t, expect, actual)
	mhs, ok := rt.middlewaresMapper[id]
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
		return func(ctx Context) {
			next(ctx)
		}
	}
	m2 := func(next HandlerFunc) HandlerFunc {
		return func(ctx Context) {
			next(ctx)
		}
	}
	m3 := func(next HandlerFunc) HandlerFunc {
		return func(ctx Context) {
			next(ctx)
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
	t.Run("when request message is nil", func(t *testing.T) {
		rt := newRouter()
		ctx := &routeContext{}
		rt.handleRequest(ctx)
	})
	t.Run("when handler and middlewares not found", func(t *testing.T) {
		rt := newRouter()
		reqMsg := NewMessage(1, []byte("test"))
		ctx := &routeContext{reqMsg: reqMsg}
		rt.handleRequest(ctx)
		assert.Nil(t, ctx.respMsg)
	})
	t.Run("when handler and middlewares found", func(t *testing.T) {
		rt := newRouter()
		var id = 1
		rt.register(id, nilHandler, func(next HandlerFunc) HandlerFunc {
			return func(ctx Context) { next(ctx) }
		})

		reqMsg := NewMessage(1, []byte("test"))
		ctx := &routeContext{reqMsg: reqMsg}
		rt.handleRequest(ctx)
		assert.Nil(t, ctx.respMsg)
	})
}

func TestRouter_wrapHandlers(t *testing.T) {
	rt := newRouter()
	t.Run("it works when there's no handler nor middleware", func(t *testing.T) {
		wrap := rt.wrapHandlers(nil, nil)
		ctx := &routeContext{}
		wrap(ctx)
		assert.Nil(t, ctx.respMsg)
	})
	t.Run("it should invoke handlers in the right order", func(t *testing.T) {
		result := make([]string, 0)

		middles := []MiddlewareFunc{
			func(next HandlerFunc) HandlerFunc {
				return func(ctx Context) {
					result = append(result, "m1-before")
					next(ctx)
				}
			},
			func(next HandlerFunc) HandlerFunc {
				return func(ctx Context) {
					result = append(result, "m2-before")
					next(ctx)
					result = append(result, "m2-after")
				}
			},
			func(next HandlerFunc) HandlerFunc {
				return func(ctx Context) {
					next(ctx)
					result = append(result, "m3-after")
				}
			},
		}
		var handler HandlerFunc = func(ctx Context) {
			result = append(result, "done")
			ctx.SetResponseMessage(NewMessage(2, []byte("done")))
		}

		wrap := rt.wrapHandlers(handler, middles)
		ctx := &routeContext{}
		wrap(ctx)
		assert.EqualValues(t, ctx.respMsg.Data(), "done")
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

		m1 := func(next HandlerFunc) HandlerFunc {
			return func(ctx Context) {
				next(ctx)
			}
		}

		m2 := func(next HandlerFunc) HandlerFunc {
			return func(ctx Context) {
				next(ctx)
			}
		}
		m3 := func(next HandlerFunc) HandlerFunc {
			return func(ctx Context) {
				next(ctx)
			}
		}
		m4 := func(next HandlerFunc) HandlerFunc {
			return func(ctx Context) {
				next(ctx)
			}
		}
		rt.registerMiddleware(m1)
		rt.registerMiddleware(m2)

		rt.register(1234, nilHandler, m3, m4)
		rt.register(12345678, nilHandler)
		rt.register(12345, nilHandler)
		rt.register(123456, nilHandler)
		rt.printHandlers("localhost")
	})
}

func TestRouter_setNotFoundHandler(t *testing.T) {
	rt := newRouter()
	assert.Nil(t, rt.notFoundHandler)
	rt.setNotFoundHandler(func(ctx Context) {})
	assert.NotNil(t, rt.notFoundHandler)
}
