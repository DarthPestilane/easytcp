package server

import (
	"github.com/DarthPestilane/easytcp/logger"
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/DarthPestilane/easytcp/router"
	"net"
	"testing"
)

//go:generate go test -bench=^BenchmarkTCPServer_\w+$ -run=none -benchmem

func BenchmarkTCPServer_NoRoute(b *testing.B) {
	logger.Log = &logger.MuteLogger{}
	s := NewTCPServer(&TCPOption{
		DontPrintRoutes: true,
	})
	go s.Serve(":0") // nolint

	<-s.accepting

	// client
	client, err := net.Dial("tcp", s.Listener.Addr().String())
	if err != nil {
		panic(err)
	}
	packedMsg, _ := s.Packer.Pack(&packet.MessageEntry{ID: 1, Data: []byte("ping")})
	for i := 0; i < b.N; i++ {
		_, _ = client.Write(packedMsg)
	}
}

func BenchmarkTCPServer_NotFoundHandler(b *testing.B) {
	logger.Log = &logger.MuteLogger{}
	s := NewTCPServer(&TCPOption{
		DontPrintRoutes: true,
	})
	s.NotFoundHandler(func(ctx *router.Context) (*packet.MessageEntry, error) {
		return ctx.Response(0, []byte("not found"))
	})
	go s.Serve(":0") // nolint

	<-s.accepting

	// client
	client, err := net.Dial("tcp", s.Listener.Addr().String())
	if err != nil {
		panic(err)
	}
	packedMsg, _ := s.Packer.Pack(&packet.MessageEntry{ID: 1, Data: []byte("ping")})
	for i := 0; i < b.N; i++ {
		_, _ = client.Write(packedMsg)
	}
}

func BenchmarkTCPServer_OneHandler(b *testing.B) {
	logger.Log = &logger.MuteLogger{}
	s := NewTCPServer(&TCPOption{
		DontPrintRoutes: true,
	})
	s.AddRoute(1, func(ctx *router.Context) (*packet.MessageEntry, error) {
		return ctx.Response(2, []byte("pong"))
	})
	go s.Serve(":0") // nolint

	<-s.accepting

	// client
	client, err := net.Dial("tcp", s.Listener.Addr().String())
	if err != nil {
		panic(err)
	}
	packedMsg, _ := s.Packer.Pack(&packet.MessageEntry{ID: 1, Data: []byte("ping")})
	for i := 0; i < b.N; i++ {
		_, _ = client.Write(packedMsg)
	}
}

func BenchmarkTCPServer_ManyHandlers(b *testing.B) {
	logger.Log = &logger.MuteLogger{}
	s := NewTCPServer(&TCPOption{
		DontPrintRoutes: true,
	})

	var m router.MiddlewareFunc = func(next router.HandlerFunc) router.HandlerFunc {
		return func(ctx *router.Context) (*packet.MessageEntry, error) {
			return next(ctx)
		}
	}

	s.AddRoute(1, func(ctx *router.Context) (*packet.MessageEntry, error) {
		return ctx.Response(2, []byte("pong"))
	}, m, m)

	go s.Serve(":0") // nolint

	<-s.accepting

	// client
	client, err := net.Dial("tcp", s.Listener.Addr().String())
	if err != nil {
		panic(err)
	}
	packedMsg, _ := s.Packer.Pack(&packet.MessageEntry{ID: 1, Data: []byte("ping")})
	for i := 0; i < b.N; i++ {
		_, _ = client.Write(packedMsg)
	}
}

func BenchmarkTCPServer_OneRouteSet(b *testing.B) {
	logger.Log = &logger.MuteLogger{}
	s := NewTCPServer(&TCPOption{
		DontPrintRoutes: true,
	})
	s.AddRoute(1, func(ctx *router.Context) (*packet.MessageEntry, error) {
		ctx.Set("key", "value")
		return ctx.Response(2, []byte("pong"))
	})
	go s.Serve(":0") // nolint

	<-s.accepting

	// client
	client, err := net.Dial("tcp", s.Listener.Addr().String())
	if err != nil {
		panic(err)
	}
	packedMsg, _ := s.Packer.Pack(&packet.MessageEntry{ID: 1, Data: []byte("ping")})
	for i := 0; i < b.N; i++ {
		_, _ = client.Write(packedMsg)
	}
}

func BenchmarkTCPServer_OneRouteJsonCodec(b *testing.B) {
	logger.Log = &logger.MuteLogger{}
	s := NewTCPServer(&TCPOption{
		Codec:           &packet.JsonCodec{},
		DontPrintRoutes: true,
	})
	s.AddRoute(1, func(ctx *router.Context) (*packet.MessageEntry, error) {
		req := make(map[string]string)
		if err := ctx.Bind(&req); err != nil {
			panic(err)
		}
		return ctx.Response(2, map[string]string{"data": "pong"})
	})
	go s.Serve(":0") // nolint

	<-s.accepting

	// client
	client, err := net.Dial("tcp", s.Listener.Addr().String())
	if err != nil {
		panic(err)
	}
	packedMsg, _ := s.Packer.Pack(&packet.MessageEntry{ID: 1, Data: []byte(`{"data": "ping"}`)})
	for i := 0; i < b.N; i++ {
		_, _ = client.Write(packedMsg)
	}
}
