package router

import (
	"context"
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/DarthPestilane/easytcp/session"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
	"time"
)

func TestRouter_Loop(t *testing.T) {
	rt := Inst()

	msg, err := (&packet.DefaultPacker{}).Pack(1, []byte("hello"))
	assert.NoError(t, err)

	t.Run("usually router can receive the request from session", func(t *testing.T) {
		r, w := net.Pipe()
		s := session.New(r, &packet.DefaultPacker{}, &packet.DefaultCodec{})
		go s.ReadLoop()
		go func() { _, _ = w.Write(msg) }()
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
		defer cancel()
		err := rt.Loop(ctx, s) // loop before session closed
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context done")
		s.Close()
		assert.NoError(t, s.WaitToClose())
	})
	t.Run("it should return error if session is closed", func(t *testing.T) {
		r, w := net.Pipe()
		s := session.New(r, &packet.DefaultPacker{}, &packet.DefaultCodec{})
		go s.ReadLoop()
		go func() { _, _ = w.Write(msg) }()
		s.Close()
		assert.NoError(t, s.WaitToClose())
		err := rt.Loop(context.Background(), s) // loop after session closed
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "receive request err: channel closed")
	})
}

func TestRouter_handleReq(t *testing.T) {
	rt := Inst()
	t.Run("it should return error when handler not found", func(t *testing.T) {
		s := session.New(nil, &packet.DefaultPacker{}, &packet.DefaultCodec{})
		req := &packet.Request{Id: 123}
		err := rt.handleReq(s, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "handler not found")
	})
	t.Run("it should return error when session's closed", func(t *testing.T) {
		rt.Register(123, func(s *session.Session, req *packet.Request) *packet.Response {
			return &packet.Response{}
		})
		s := session.New(nil, &packet.DefaultPacker{}, &packet.DefaultCodec{})
		s.Close()
		req := &packet.Request{Id: 123}
		err := rt.handleReq(s, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "session send response err")
	})
	t.Run("it should be ok when handler returns nil response", func(t *testing.T) {
		rt.Register(123, func(s *session.Session, req *packet.Request) *packet.Response {
			return nil
		})
		s := session.New(nil, &packet.DefaultPacker{}, &packet.DefaultCodec{})
		req := &packet.Request{Id: 123}
		err := rt.handleReq(s, req)
		assert.NoError(t, err)
	})
	t.Run("it should be ok when everything's fine", func(t *testing.T) {
		rt.Register(1, func(s *session.Session, req *packet.Request) *packet.Response {
			return &packet.Response{
				Id:   2,
				Data: "world",
			}
		})
		s := session.New(nil, &packet.DefaultPacker{}, &packet.DefaultCodec{})
		req := &packet.Request{Id: 1, RawData: []byte("hello")}
		err := rt.handleReq(s, req)
		assert.NoError(t, err)
	})
}
