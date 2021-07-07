package easytcp

import (
	"github.com/DarthPestilane/easytcp/message"
	"net"
	"testing"
)

// go test -bench='^BenchmarkTCPServer_\w+$' -run=none -benchmem

func BenchmarkTCPServer_NoRoute(b *testing.B) {
	Log = &MuteLogger{}
	s := NewServer(&ServerOption{
		DontPrintRoutes: true,
	})
	go s.Serve(":0") // nolint

	<-s.accepting

	// client
	client, err := net.Dial("tcp", s.Listener.Addr().String())
	if err != nil {
		panic(err)
	}
	defer client.Close() // nolint
	packedMsg, _ := s.Packer.Pack(&message.Entry{ID: 1, Data: []byte("ping")})
	for i := 0; i < b.N; i++ {
		_, _ = client.Write(packedMsg)
	}
}

func BenchmarkTCPServer_NotFoundHandler(b *testing.B) {
	Log = &MuteLogger{}
	s := NewServer(&ServerOption{
		DontPrintRoutes: true,
	})
	s.NotFoundHandler(func(ctx *Context) (*message.Entry, error) {
		return ctx.Response(0, []byte("not found"))
	})
	go s.Serve(":0") // nolint

	<-s.accepting

	// client
	client, err := net.Dial("tcp", s.Listener.Addr().String())
	if err != nil {
		panic(err)
	}
	defer client.Close() // nolint

	packedMsg, _ := s.Packer.Pack(&message.Entry{ID: 1, Data: []byte("ping")})
	for i := 0; i < b.N; i++ {
		_, _ = client.Write(packedMsg)
	}
}

func BenchmarkTCPServer_OneHandler(b *testing.B) {
	Log = &MuteLogger{}
	s := NewServer(&ServerOption{
		DontPrintRoutes: true,
	})
	s.AddRoute(1, func(ctx *Context) (*message.Entry, error) {
		return ctx.Response(2, []byte("pong"))
	})
	go s.Serve(":0") // nolint

	<-s.accepting

	// client
	client, err := net.Dial("tcp", s.Listener.Addr().String())
	if err != nil {
		panic(err)
	}
	defer client.Close() // nolint

	packedMsg, _ := s.Packer.Pack(&message.Entry{ID: 1, Data: []byte("ping")})
	for i := 0; i < b.N; i++ {
		_, _ = client.Write(packedMsg)
	}
}

func BenchmarkTCPServer_ManyHandlers(b *testing.B) {
	Log = &MuteLogger{}
	s := NewServer(&ServerOption{
		DontPrintRoutes: true,
	})

	var m MiddlewareFunc = func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) (*message.Entry, error) {
			return next(ctx)
		}
	}

	s.AddRoute(1, func(ctx *Context) (*message.Entry, error) {
		return ctx.Response(2, []byte("pong"))
	}, m, m)

	go s.Serve(":0") // nolint

	<-s.accepting

	// client
	client, err := net.Dial("tcp", s.Listener.Addr().String())
	if err != nil {
		panic(err)
	}
	defer client.Close() // nolint

	packedMsg, _ := s.Packer.Pack(&message.Entry{ID: 1, Data: []byte("ping")})
	for i := 0; i < b.N; i++ {
		_, _ = client.Write(packedMsg)
	}
}

func BenchmarkTCPServer_OneRouteSet(b *testing.B) {
	Log = &MuteLogger{}
	s := NewServer(&ServerOption{
		DontPrintRoutes: true,
	})
	s.AddRoute(1, func(ctx *Context) (*message.Entry, error) {
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
	defer client.Close() // nolint

	packedMsg, _ := s.Packer.Pack(&message.Entry{ID: 1, Data: []byte("ping")})
	for i := 0; i < b.N; i++ {
		_, _ = client.Write(packedMsg)
	}
}

func BenchmarkTCPServer_OneRouteJsonCodec(b *testing.B) {
	Log = &MuteLogger{}
	s := NewServer(&ServerOption{
		Codec:           &JsonCodec{},
		DontPrintRoutes: true,
	})
	s.AddRoute(1, func(ctx *Context) (*message.Entry, error) {
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
	defer client.Close() // nolint

	packedMsg, _ := s.Packer.Pack(&message.Entry{ID: 1, Data: []byte(`{"data": "ping"}`)})
	for i := 0; i < b.N; i++ {
		_, _ = client.Write(packedMsg)
	}
}
