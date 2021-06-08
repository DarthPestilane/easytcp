package router

import (
	"fmt"
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/DarthPestilane/easytcp/session"
	"github.com/DarthPestilane/easytcp/session/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"reflect"
	"runtime"
	"testing"
)

func TestNew(t *testing.T) {
	rt := New()
	assert.NotNil(t, rt.log)
	assert.NotNil(t, rt.globalMiddlewares)
}

func TestRouter_Loop(t *testing.T) {
	t.Run("when session is closed", func(t *testing.T) {
		rt := New()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		sess := mock.NewMockSession(ctrl)
		reqCh := make(chan *packet.Request)
		close(reqCh)
		sess.EXPECT().RecvReq().Return(reqCh)
		sess.EXPECT().ID().Times(2).Return("test-session-id")
		rt.Loop(sess) // should return
	})
	t.Run("when received a nil request", func(t *testing.T) {
		rt := New()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		reqCh := make(chan *packet.Request)
		go func() {
			reqCh <- nil
			close(reqCh)
		}()
		sess := mock.NewMockSession(ctrl)
		sess.EXPECT().RecvReq().Times(2).Return(reqCh)
		sess.EXPECT().ID().Times(2).Return("test session id")
		loopDone := make(chan struct{})
		go func() {
			rt.Loop(sess) // should not call to handler
			close(loopDone)
		}()
		<-loopDone
	})
	t.Run("when received a non-nil request", func(t *testing.T) {
		t.Run("when handler returns error", func(t *testing.T) {
			rt := New()

			rt.Register(1, func(s session.Session, req *packet.Request) (*packet.Response, error) {
				assert.EqualValues(t, req.Id, 1)
				assert.EqualValues(t, req.RawSize, 4)
				assert.Equal(t, req.RawData, []byte("test"))
				return nil, fmt.Errorf("some err")
			})
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			reqCh := make(chan *packet.Request)
			go func() {
				reqCh <- &packet.Request{Id: 1, RawData: []byte("test"), RawSize: 4}
				close(reqCh)
			}()
			sess := mock.NewMockSession(ctrl)
			sess.EXPECT().RecvReq().Times(2).Return(reqCh)
			sess.EXPECT().ID().MaxTimes(3).Return("test session id")
			loopDone := make(chan struct{})
			go func() {
				rt.Loop(sess) // should not call to handler
				close(loopDone)
			}()
			<-loopDone
		})
		t.Run("when handler returns no error", func(t *testing.T) {
			rt := New()

			rt.Register(1, defaultHandler)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			reqCh := make(chan *packet.Request)
			go func() {
				reqCh <- &packet.Request{Id: 1, RawData: []byte("test"), RawSize: 4}
				close(reqCh)
			}()
			sess := mock.NewMockSession(ctrl)
			sess.EXPECT().RecvReq().Times(2).Return(reqCh)
			sess.EXPECT().ID().Times(2).Return("test session id")
			loopDone := make(chan struct{})
			go func() {
				rt.Loop(sess) // should not call to handler
				close(loopDone)
			}()
			<-loopDone
		})
	})
}

func TestRouter_Register(t *testing.T) {
	rt := New()

	var id uint = 1

	rt.Register(id, nil)
	_, ok := rt.handlerMapper.Load(id)
	assert.False(t, ok)
	_, ok = rt.middlewaresMapper.Load(id)
	assert.False(t, ok)

	h := defaultHandler
	m1 := func(next HandlerFunc) HandlerFunc {
		return func(s session.Session, req *packet.Request) (*packet.Response, error) {
			return next(s, req)
		}
	}
	m2 := func(next HandlerFunc) HandlerFunc {
		return func(s session.Session, req *packet.Request) (*packet.Response, error) {
			return next(s, req)
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
	rt := New()

	rt.RegisterMiddleware()
	assert.Len(t, rt.globalMiddlewares, 0)

	rt.RegisterMiddleware(nil, nil)
	assert.Len(t, rt.globalMiddlewares, 0)

	m1 := func(next HandlerFunc) HandlerFunc {
		return func(s session.Session, req *packet.Request) (*packet.Response, error) {
			return next(s, req)
		}
	}
	m2 := func(next HandlerFunc) HandlerFunc {
		return func(s session.Session, req *packet.Request) (*packet.Response, error) {
			return next(s, req)
		}
	}
	m3 := func(next HandlerFunc) HandlerFunc {
		return func(s session.Session, req *packet.Request) (*packet.Response, error) {
			return next(s, req)
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
		rt := New()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		sess := mock.NewMockSession(ctrl)
		assert.Nil(t, rt.handleReq(sess, &packet.Request{Id: 1}))
	})
	t.Run("when handler and middlewares found", func(t *testing.T) {
		rt := New()
		var id uint = 1
		rt.Register(id, defaultHandler, func(next HandlerFunc) HandlerFunc {
			return func(s session.Session, req *packet.Request) (*packet.Response, error) { return next(s, req) }
		})

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		sess := mock.NewMockSession(ctrl)
		assert.Nil(t, rt.handleReq(sess, &packet.Request{Id: id}))
	})
	t.Run("when handler returns error", func(t *testing.T) {
		rt := New()
		var id uint = 1
		rt.Register(id, func(s session.Session, req *packet.Request) (*packet.Response, error) {
			return nil, fmt.Errorf("some err")
		})

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		sess := mock.NewMockSession(ctrl)
		err := rt.handleReq(sess, &packet.Request{Id: id})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "handler err")
	})
	t.Run("when handler returns a non-nil response", func(t *testing.T) {
		t.Run("when session send resp failed", func(t *testing.T) {
			rt := New()
			var id uint = 1

			resp := &packet.Response{}
			rt.Register(id, func(s session.Session, req *packet.Request) (*packet.Response, error) {
				return resp, nil
			})

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			sess := mock.NewMockSession(ctrl)
			sess.EXPECT().SendResp(resp).Return(false, fmt.Errorf("some err"))
			err := rt.handleReq(sess, &packet.Request{Id: id})
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "session send response err")
		})
		t.Run("when session send resp without error", func(t *testing.T) {
			rt := New()
			var id uint = 1

			resp := &packet.Response{}
			rt.Register(id, func(s session.Session, req *packet.Request) (*packet.Response, error) {
				return resp, nil
			})

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			sess := mock.NewMockSession(ctrl)
			sess.EXPECT().SendResp(resp).Return(false, nil)
			err := rt.handleReq(sess, &packet.Request{Id: id})
			assert.NoError(t, err)
		})
	})
}

func TestRouter_wrapHandlers(t *testing.T) {
	rt := New()
	t.Run("it works when there's no handler nor middleware", func(t *testing.T) {
		wrap := rt.wrapHandlers(nil, nil)
		resp, err := wrap(nil, nil)
		assert.NoError(t, err)
		assert.Nil(t, resp)
	})
	t.Run("it should invoke handlers in the right order", func(t *testing.T) {
		result := make([]string, 0)

		middles := []MiddlewareFunc{
			func(next HandlerFunc) HandlerFunc {
				return func(s session.Session, req *packet.Request) (*packet.Response, error) {
					result = append(result, "m1-before")
					return next(s, req)
				}
			},
			func(next HandlerFunc) HandlerFunc {
				return func(s session.Session, req *packet.Request) (*packet.Response, error) {
					result = append(result, "m2-before")
					resp, err := next(s, req)
					result = append(result, "m2-after")
					return resp, err
				}
			},
			func(next HandlerFunc) HandlerFunc {
				return func(s session.Session, req *packet.Request) (*packet.Response, error) {
					resp, err := next(s, req)
					result = append(result, "m3-after")
					return resp, err
				}
			},
		}
		var handler HandlerFunc = func(s session.Session, req *packet.Request) (*packet.Response, error) {
			result = append(result, "done")
			return &packet.Response{Data: "done"}, nil
		}

		wrap := rt.wrapHandlers(handler, middles)
		resp, err := wrap(nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, resp.Data, "done")
		assert.Equal(t, result, []string{"m1-before", "m2-before", "done", "m3-after", "m2-after"})
	})
}
