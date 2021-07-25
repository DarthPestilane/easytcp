package easytcp

import (
	"github.com/DarthPestilane/easytcp/message"
	"github.com/DarthPestilane/easytcp/test_data/pb"
	"net"
	"testing"
)

// go test -bench="^BenchmarkTCPServer_\w+$" -run=none -benchmem

func Benchmark_NoRoute(b *testing.B) {
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
	packedMsg, _ := s.Packer.Pack(&message.Entry{ID: uint32(1), Data: []byte("ping")})
	beforeBench(b)
	for i := 0; i < b.N; i++ {
		_, _ = client.Write(packedMsg)
	}
}

func Benchmark_NotFoundHandler(b *testing.B) {
	Log = &MuteLogger{}
	s := NewServer(&ServerOption{
		DontPrintRoutes: true,
	})
	s.NotFoundHandler(func(ctx *Context) (*message.Entry, error) {
		return ctx.Response(uint32(0), []byte("not found"))
	})
	go s.Serve(":0") // nolint

	<-s.accepting

	// client
	client, err := net.Dial("tcp", s.Listener.Addr().String())
	if err != nil {
		panic(err)
	}
	defer client.Close() // nolint

	packedMsg, _ := s.Packer.Pack(&message.Entry{ID: uint32(1), Data: []byte("ping")})
	beforeBench(b)
	for i := 0; i < b.N; i++ {
		_, _ = client.Write(packedMsg)
	}
}

func Benchmark_OneHandler(b *testing.B) {
	Log = &MuteLogger{}
	s := NewServer(&ServerOption{
		DontPrintRoutes: true,
	})
	s.AddRoute(uint32(1), func(ctx *Context) (*message.Entry, error) {
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

	packedMsg, _ := s.Packer.Pack(&message.Entry{ID: uint32(1), Data: []byte("ping")})
	beforeBench(b)
	for i := 0; i < b.N; i++ {
		_, _ = client.Write(packedMsg)
	}
}

func Benchmark_ManyHandlers(b *testing.B) {
	Log = &MuteLogger{}
	s := NewServer(&ServerOption{
		DontPrintRoutes: true,
	})

	var m MiddlewareFunc = func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) (*message.Entry, error) {
			return next(ctx)
		}
	}

	s.AddRoute(uint32(1), func(ctx *Context) (*message.Entry, error) {
		return ctx.Response(uint32(2), []byte("pong"))
	}, m, m)

	go s.Serve(":0") // nolint

	<-s.accepting

	// client
	client, err := net.Dial("tcp", s.Listener.Addr().String())
	if err != nil {
		panic(err)
	}
	defer client.Close() // nolint

	packedMsg, _ := s.Packer.Pack(&message.Entry{ID: uint32(1), Data: []byte("ping")})
	beforeBench(b)
	for i := 0; i < b.N; i++ {
		_, _ = client.Write(packedMsg)
	}
}

func Benchmark_OneRouteSet(b *testing.B) {
	Log = &MuteLogger{}
	s := NewServer(&ServerOption{
		DontPrintRoutes: true,
	})
	s.AddRoute(uint32(1), func(ctx *Context) (*message.Entry, error) {
		ctx.Set("key", "value")
		v, _ := ctx.Get("key")
		return ctx.Response(uint32(2), []byte(v.(string)))
	})
	go s.Serve(":0") // nolint

	<-s.accepting

	// client
	client, err := net.Dial("tcp", s.Listener.Addr().String())
	if err != nil {
		panic(err)
	}
	defer client.Close() // nolint

	packedMsg, _ := s.Packer.Pack(&message.Entry{ID: uint32(1), Data: []byte("ping")})
	beforeBench(b)
	for i := 0; i < b.N; i++ {
		_, _ = client.Write(packedMsg)
	}
}

func Benchmark_OneRouteJsonCodec(b *testing.B) {
	Log = &MuteLogger{}
	s := NewServer(&ServerOption{
		Codec:           &JsonCodec{},
		DontPrintRoutes: true,
	})
	s.AddRoute(uint32(1), func(ctx *Context) (*message.Entry, error) {
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

	packedMsg, _ := s.Packer.Pack(&message.Entry{ID: uint32(1), Data: []byte(`{"data": "ping"}`)})
	beforeBench(b)
	for i := 0; i < b.N; i++ {
		_, _ = client.Write(packedMsg)
	}
}

func Benchmark_OneRouteProtobufCodec(b *testing.B) {
	Log = &MuteLogger{}
	s := NewServer(&ServerOption{
		Codec:           &ProtobufCodec{},
		DontPrintRoutes: true,
	})
	s.AddRoute(uint32(1), func(ctx *Context) (*message.Entry, error) {
		var req pb.Sample
		if err := ctx.Bind(&req); err != nil {
			panic(err)
		}
		return ctx.Response(2, &pb.Sample{Foo: "test-resp", Bar: req.Bar + 1})
	})
	go s.Serve(":0") // nolint

	<-s.accepting

	// client
	client, err := net.Dial("tcp", s.Listener.Addr().String())
	if err != nil {
		panic(err)
	}
	defer client.Close() // nolint

	data, _ := s.Codec.Encode(&pb.Sample{Foo: "test", Bar: 1})
	packedMsg, _ := s.Packer.Pack(&message.Entry{ID: uint32(1), Data: data})
	beforeBench(b)
	for i := 0; i < b.N; i++ {
		_, _ = client.Write(packedMsg)
	}
}

func beforeBench(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
}
