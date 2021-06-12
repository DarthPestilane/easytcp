package router

import (
	"fmt"
	"github.com/DarthPestilane/easytcp/packet"
	mockPacker "github.com/DarthPestilane/easytcp/packet/mock"
	"github.com/DarthPestilane/easytcp/session/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"reflect"
	"runtime"
	"testing"
)

func TestNewRouter(t *testing.T) {
	rt := NewRouter()
	assert.NotNil(t, rt.log)
	assert.NotNil(t, rt.globalMiddlewares)
}

func TestRouter_RouteLoop(t *testing.T) {
	t.Run("when session is closed", func(t *testing.T) {
		rt := NewRouter()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		sess := mock.NewMockSession(ctrl)
		reqCh := make(chan packet.Message)
		close(reqCh)
		sess.EXPECT().RecvReq().Return(reqCh)
		sess.EXPECT().ID().Times(2).Return("test-session-id")
		rt.RouteLoop(sess) // should return
	})
	t.Run("when received a nil request", func(t *testing.T) {
		rt := NewRouter()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		reqCh := make(chan packet.Message)
		go func() {
			reqCh <- nil
			close(reqCh)
		}()
		sess := mock.NewMockSession(ctrl)
		sess.EXPECT().RecvReq().Times(2).Return(reqCh)
		sess.EXPECT().ID().Times(2).Return("test session id")
		loopDone := make(chan struct{})
		go func() {
			rt.RouteLoop(sess) // should not call to handler
			close(loopDone)
		}()
		<-loopDone
	})
	t.Run("when received a non-nil request", func(t *testing.T) {
		t.Run("when handler returns error", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			msg := mockPacker.NewMockMessage(ctrl)
			msg.EXPECT().GetID().AnyTimes().Return(uint(1))
			msg.EXPECT().GetData().AnyTimes().Return([]byte("test"))
			msg.EXPECT().GetSize().AnyTimes().Return(uint(4))

			rt := NewRouter()

			rt.Register(1, func(ctx *Context) (packet.Message, error) {
				assert.EqualValues(t, ctx.MsgID(), 1)
				assert.EqualValues(t, ctx.MsgSize(), 4)
				assert.Equal(t, ctx.MsgRawData(), []byte("test"))
				return nil, fmt.Errorf("some err")
			})

			reqCh := make(chan packet.Message)
			go func() {
				reqCh <- msg
				close(reqCh)
			}()
			sess := mock.NewMockSession(ctrl)
			sess.EXPECT().RecvReq().Times(2).Return(reqCh)
			sess.EXPECT().ID().MaxTimes(3).Return("test session id")
			loopDone := make(chan struct{})
			go func() {
				rt.RouteLoop(sess) // should not call to handler
				close(loopDone)
			}()
			<-loopDone
		})
		t.Run("when handler returns no error", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			msg := mockPacker.NewMockMessage(ctrl)
			msg.EXPECT().GetID().AnyTimes().Return(uint(1))

			rt := NewRouter()

			rt.Register(1, defaultHandler)

			reqCh := make(chan packet.Message)
			go func() {
				reqCh <- msg
				close(reqCh)
			}()
			sess := mock.NewMockSession(ctrl)
			sess.EXPECT().RecvReq().Times(2).Return(reqCh)
			sess.EXPECT().ID().Times(2).Return("test session id")
			loopDone := make(chan struct{})
			go func() {
				rt.RouteLoop(sess) // should not call to handler
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

	h := defaultHandler
	m1 := func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) (packet.Message, error) {
			return next(ctx)
		}
	}
	m2 := func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) (packet.Message, error) {
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
		return func(ctx *Context) (packet.Message, error) {
			return next(ctx)
		}
	}
	m2 := func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) (packet.Message, error) {
			return next(ctx)
		}
	}
	m3 := func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) (packet.Message, error) {
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

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		msg := mockPacker.NewMockMessage(ctrl)
		msg.EXPECT().GetID().AnyTimes().Return(uint(1))
		sess := mock.NewMockSession(ctrl)

		assert.Nil(t, rt.handleReq(sess, msg))
	})
	t.Run("when handler and middlewares found", func(t *testing.T) {
		rt := NewRouter()
		var id uint = 1
		rt.Register(id, defaultHandler, func(next HandlerFunc) HandlerFunc {
			return func(ctx *Context) (packet.Message, error) { return next(ctx) }
		})

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		sess := mock.NewMockSession(ctrl)
		msg := mockPacker.NewMockMessage(ctrl)
		msg.EXPECT().GetID().AnyTimes().Return(id)

		assert.Nil(t, rt.handleReq(sess, msg))
	})
	t.Run("when handler returns error", func(t *testing.T) {
		rt := NewRouter()
		var id uint = 1
		rt.Register(id, func(ctx *Context) (packet.Message, error) {
			return nil, fmt.Errorf("some err")
		})

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		sess := mock.NewMockSession(ctrl)
		msg := mockPacker.NewMockMessage(ctrl)
		msg.EXPECT().GetID().AnyTimes().Return(id)

		err := rt.handleReq(sess, msg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "handler err")
	})
	t.Run("when handler returns a non-nil response", func(t *testing.T) {
		t.Run("when session send resp failed", func(t *testing.T) {
			var id uint = 1

			rt := NewRouter()

			// register route
			rt.Register(id, func(ctx *Context) (packet.Message, error) {
				return &packet.DefaultMsg{}, nil
			})

			// mock
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			message := mockPacker.NewMockMessage(ctrl)
			message.EXPECT().GetID().AnyTimes().Return(id)

			sess := mock.NewMockSession(ctrl)
			sess.EXPECT().SendResp(gomock.Any()).Return(fmt.Errorf("some err"))

			err := rt.handleReq(sess, message)
			assert.Error(t, err)
		})
		t.Run("when session send resp without error", func(t *testing.T) {
			rt := NewRouter()
			var id uint = 1

			rt.Register(id, func(ctx *Context) (packet.Message, error) {
				return &packet.DefaultMsg{}, nil
			})

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			message := mockPacker.NewMockMessage(ctrl)
			message.EXPECT().GetID().AnyTimes().Return(id)

			sess := mock.NewMockSession(ctrl)
			sess.EXPECT().SendResp(gomock.Any()).Return(nil)

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
				return func(ctx *Context) (packet.Message, error) {
					result = append(result, "m1-before")
					return next(ctx)
				}
			},
			func(next HandlerFunc) HandlerFunc {
				return func(ctx *Context) (packet.Message, error) {
					result = append(result, "m2-before")
					resp, err := next(ctx)
					result = append(result, "m2-after")
					return resp, err
				}
			},
			func(next HandlerFunc) HandlerFunc {
				return func(ctx *Context) (packet.Message, error) {
					resp, err := next(ctx)
					result = append(result, "m3-after")
					return resp, err
				}
			},
		}
		var handler HandlerFunc = func(ctx *Context) (packet.Message, error) {
			result = append(result, "done")
			msg := &packet.DefaultMsg{}
			msg.Setup(2, []byte("done"))
			return msg, nil
		}

		wrap := rt.wrapHandlers(handler, middles)
		resp, err := wrap(nil)
		assert.NoError(t, err)
		assert.EqualValues(t, resp.GetData(), "done")
		assert.Equal(t, result, []string{"m1-before", "m2-before", "done", "m3-after", "m2-after"})
	})
}
