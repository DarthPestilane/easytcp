package router

import (
	"fmt"
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/DarthPestilane/easytcp/session"
	"github.com/stretchr/testify/assert"
	"net"
	"reflect"
	"runtime"
	"testing"
	"time"
)

func TestRouter_Loop(t *testing.T) {
	rt := Instance()

	msg, err := (&packet.DefaultPacker{}).Pack(1, []byte("hello"))
	assert.NoError(t, err)

	t.Run("usually router can receive the request from session", func(t *testing.T) {
		r, w := net.Pipe()
		s := session.New(r, &packet.DefaultPacker{}, &packet.StringCodec{})
		go s.ReadLoop()
		go func() { _, _ = w.Write(msg) }()
		go func() {
			err := rt.Loop(s)
			assert.Error(t, err) // loop before session close
			assert.Contains(t, err.Error(), "channel closed")
		}()
		time.After(time.Millisecond * 10)
		s.Close()
		assert.NoError(t, s.WaitToClose())
	})
	t.Run("it should return error if session is closed", func(t *testing.T) {
		r, w := net.Pipe()
		s := session.New(r, &packet.DefaultPacker{}, &packet.StringCodec{})
		go s.ReadLoop()
		go func() { _, _ = w.Write(msg) }()
		s.Close()
		assert.NoError(t, s.WaitToClose())
		err := rt.Loop(s) // loop after session closed
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "receive request err: channel closed")
	})
}

func TestRouter_handleReq(t *testing.T) {
	rt := Instance()
	t.Run("it should be ok when handler not found", func(t *testing.T) {
		s := session.New(nil, &packet.DefaultPacker{}, &packet.StringCodec{})
		req := &packet.Request{Id: 123}
		err := rt.handleReq(s, req)
		assert.NoError(t, err)
	})
	t.Run("it should return error when session's closed", func(t *testing.T) {
		rt.Register(123, func(s *session.Session, req *packet.Request) (*packet.Response, error) {
			return &packet.Response{}, nil
		})
		s := session.New(nil, &packet.DefaultPacker{}, &packet.StringCodec{})
		s.Close()
		req := &packet.Request{Id: 123}
		err := rt.handleReq(s, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "session send response err")
	})
	t.Run("it should be ok when handler returns nil response", func(t *testing.T) {
		rt.Register(123, func(s *session.Session, req *packet.Request) (*packet.Response, error) {
			return nil, nil
		})
		s := session.New(nil, &packet.DefaultPacker{}, &packet.StringCodec{})
		req := &packet.Request{Id: 123}
		err := rt.handleReq(s, req)
		assert.NoError(t, err)
	})
	t.Run("it should return error when handler returns error", func(t *testing.T) {
		rt.Register(123, func(s *session.Session, req *packet.Request) (*packet.Response, error) {
			return nil, fmt.Errorf("some error")
		})
		s := session.New(nil, &packet.DefaultPacker{}, &packet.StringCodec{})
		req := &packet.Request{Id: 123}
		err := rt.handleReq(s, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "handler err")
	})
	t.Run("it should be ok when everything's fine", func(t *testing.T) {
		rt.Register(1, func(s *session.Session, req *packet.Request) (*packet.Response, error) {
			return &packet.Response{
				Id:   2,
				Data: "world",
			}, nil
		})
		s := session.New(nil, &packet.DefaultPacker{}, &packet.StringCodec{})
		req := &packet.Request{Id: 1, RawData: []byte("hello")}
		err := rt.handleReq(s, req)
		assert.NoError(t, err)
	})
}

func TestRouter_wrapHandlers(t *testing.T) {
	rt := Instance()
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
				return func(s *session.Session, req *packet.Request) (*packet.Response, error) {
					result = append(result, "m1-before")
					return next(s, req)
				}
			},
			func(next HandlerFunc) HandlerFunc {
				return func(s *session.Session, req *packet.Request) (*packet.Response, error) {
					result = append(result, "m2-before")
					resp, err := next(s, req)
					result = append(result, "m2-after")
					return resp, err
				}
			},
			func(next HandlerFunc) HandlerFunc {
				return func(s *session.Session, req *packet.Request) (*packet.Response, error) {
					resp, err := next(s, req)
					result = append(result, "m3-after")
					return resp, err
				}
			},
		}
		var handler HandlerFunc = func(s *session.Session, req *packet.Request) (*packet.Response, error) {
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

func TestRouter_RegisterMiddleware(t *testing.T) {
	rt := Instance()

	var middle01 MiddlewareFunc = func(next HandlerFunc) HandlerFunc {
		return func(s *session.Session, req *packet.Request) (*packet.Response, error) {
			return nil, nil
		}
	}
	var middle02 MiddlewareFunc = func(next HandlerFunc) HandlerFunc {
		return func(s *session.Session, req *packet.Request) (*packet.Response, error) {
			return nil, nil
		}
	}

	rt.RegisterMiddleware()
	assert.Len(t, rt.globalMiddlewares, 0)

	rt.RegisterMiddleware(middle01)
	assert.Len(t, rt.globalMiddlewares, 1)

	rt.RegisterMiddleware(middle01, middle02)
	assert.Len(t, rt.globalMiddlewares, 3)

	expects := []MiddlewareFunc{middle01, middle01, middle02}
	for i, v := range rt.globalMiddlewares {
		expect := runtime.FuncForPC(reflect.ValueOf(expects[i]).Pointer()).Name()
		actual := runtime.FuncForPC(reflect.ValueOf(v).Pointer()).Name()
		assert.Equal(t, expect, actual)
	}
}
