package easytcp

import (
	"github.com/DarthPestilane/easytcp/internal/test_data/msgpack"
	"github.com/DarthPestilane/easytcp/internal/test_data/pb"
	"github.com/DarthPestilane/easytcp/message"
	"net"
	"testing"
)

// go test -bench="^BenchmarkTCPServer_\w+$" -run=none -benchmem -benchtime=250000x

type mutedLogger struct{}

func (m *mutedLogger) Errorf(_ string, _ ...interface{}) {}
func (m *mutedLogger) Tracef(_ string, _ ...interface{}) {}

func Benchmark_NoRoute(b *testing.B) {
	s := NewServer(&ServerOption{
		DoNotPrintRoutes: true,
	})
	go s.Serve(":0") // nolint
	defer s.Stop()   // nolint

	<-s.accepting

	// client
	client, err := net.Dial("tcp", s.Listener.Addr().String())
	if err != nil {
		panic(err)
	}
	defer client.Close() // nolint

	packedMsg, _ := s.Packer.Pack(&message.Entry{ID: 1, Data: []byte("ping")})
	beforeBench(b)
	for i := 0; i < b.N; i++ {
		_, _ = client.Write(packedMsg)
	}
}

func Benchmark_NotFoundHandler(b *testing.B) {
	s := NewServer(&ServerOption{
		DoNotPrintRoutes: true,
	})
	s.NotFoundHandler(func(ctx *Context) error {
		return ctx.Response(0, []byte("not found"))
	})
	go s.Serve(":0") // nolint
	defer s.Stop()   // nolint

	<-s.accepting

	// client
	client, err := net.Dial("tcp", s.Listener.Addr().String())
	if err != nil {
		panic(err)
	}
	defer client.Close() // nolint

	packedMsg, _ := s.Packer.Pack(&message.Entry{ID: 1, Data: []byte("ping")})
	beforeBench(b)
	for i := 0; i < b.N; i++ {
		_, _ = client.Write(packedMsg)
	}
}

func Benchmark_OneHandler(b *testing.B) {
	s := NewServer(&ServerOption{
		DoNotPrintRoutes: true,
	})
	s.AddRoute(1, func(ctx *Context) error {
		return ctx.Response(2, []byte("pong"))
	})
	go s.Serve(":0") // nolint
	defer s.Stop()   // nolint

	<-s.accepting

	// client
	client, err := net.Dial("tcp", s.Listener.Addr().String())
	if err != nil {
		panic(err)
	}
	defer client.Close() // nolint

	packedMsg, _ := s.Packer.Pack(&message.Entry{ID: 1, Data: []byte("ping")})
	beforeBench(b)
	for i := 0; i < b.N; i++ {
		_, _ = client.Write(packedMsg)
	}
}

func Benchmark_ManyHandlers(b *testing.B) {
	s := NewServer(&ServerOption{
		DoNotPrintRoutes: true,
	})

	var m MiddlewareFunc = func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			return next(ctx)
		}
	}

	s.AddRoute(1, func(ctx *Context) error {
		return ctx.Response(2, []byte("pong"))
	}, m, m)

	go s.Serve(":0") // nolint
	defer s.Stop()   // nolint

	<-s.accepting

	// client
	client, err := net.Dial("tcp", s.Listener.Addr().String())
	if err != nil {
		panic(err)
	}
	defer client.Close() // nolint

	packedMsg, _ := s.Packer.Pack(&message.Entry{ID: 1, Data: []byte("ping")})
	beforeBench(b)
	for i := 0; i < b.N; i++ {
		_, _ = client.Write(packedMsg)
	}
}

func Benchmark_OneRouteCtxGetSet(b *testing.B) {
	s := NewServer(&ServerOption{
		DoNotPrintRoutes: true,
	})
	s.AddRoute(1, func(ctx *Context) error {
		ctx.Set("key", "value")
		v := ctx.MustGet("key").(string)
		return ctx.Response(2, []byte(v))
	})
	go s.Serve(":0") // nolint
	defer s.Stop()   // nolint

	<-s.accepting

	// client
	client, err := net.Dial("tcp", s.Listener.Addr().String())
	if err != nil {
		panic(err)
	}
	defer client.Close() // nolint

	packedMsg, _ := s.Packer.Pack(&message.Entry{ID: 1, Data: []byte("ping")})
	beforeBench(b)
	for i := 0; i < b.N; i++ {
		_, _ = client.Write(packedMsg)
	}
}

func Benchmark_OneRouteMessageGetSet(b *testing.B) {
	s := NewServer(&ServerOption{
		DoNotPrintRoutes: true,
	})
	s.AddRoute(1, func(ctx *Context) error {
		ctx.Message().Set("key", []byte("val"))
		v := ctx.Message().MustGet("key").([]byte)
		return ctx.Response(2, v)
	})
	go s.Serve(":0") // nolint
	defer s.Stop()   // nolint

	<-s.accepting

	// client
	client, err := net.Dial("tcp", s.Listener.Addr().String())
	if err != nil {
		panic(err)
	}
	defer client.Close() // nolint

	packedMsg, _ := s.Packer.Pack(&message.Entry{ID: 1, Data: []byte("ping")})
	beforeBench(b)
	for i := 0; i < b.N; i++ {
		_, _ = client.Write(packedMsg)
	}
}

func Benchmark_OneRouteJsonCodec(b *testing.B) {
	s := NewServer(&ServerOption{
		Codec:            &JsonCodec{},
		DoNotPrintRoutes: true,
	})
	s.AddRoute(1, func(ctx *Context) error {
		req := make(map[string]string)
		ctx.MustBind(&req)
		return ctx.Response(2, map[string]string{"data": "pong"})
	})
	go s.Serve(":0") // nolint
	defer s.Stop()   // nolint

	<-s.accepting

	// client
	client, err := net.Dial("tcp", s.Listener.Addr().String())
	if err != nil {
		panic(err)
	}
	defer client.Close() // nolint

	packedMsg, _ := s.Packer.Pack(&message.Entry{ID: 1, Data: []byte(`{"data": "ping"}`)})
	beforeBench(b)
	for i := 0; i < b.N; i++ {
		_, _ = client.Write(packedMsg)
	}
}

func Benchmark_OneRouteProtobufCodec(b *testing.B) {
	s := NewServer(&ServerOption{
		Codec:            &ProtobufCodec{},
		DoNotPrintRoutes: true,
	})
	s.AddRoute(1, func(ctx *Context) error {
		var req pb.Sample
		ctx.MustBind(&req)
		return ctx.Response(2, &pb.Sample{Foo: "test-resp", Bar: req.Bar + 1})
	})
	go s.Serve(":0") // nolint
	defer s.Stop()   // nolint

	<-s.accepting

	// client
	client, err := net.Dial("tcp", s.Listener.Addr().String())
	if err != nil {
		panic(err)
	}
	defer client.Close() // nolint

	data, _ := s.Codec.Encode(&pb.Sample{Foo: "test", Bar: 1})
	packedMsg, _ := s.Packer.Pack(&message.Entry{ID: 1, Data: data})
	beforeBench(b)
	for i := 0; i < b.N; i++ {
		_, _ = client.Write(packedMsg)
	}
}

func Benchmark_OneRouteMsgpackCodec(b *testing.B) {
	s := NewServer(&ServerOption{
		Codec:            &MsgpackCodec{},
		DoNotPrintRoutes: true,
	})
	s.AddRoute(1, func(ctx *Context) error {
		var req msgpack.Sample
		ctx.MustBind(&req)
		return ctx.Response(2, &msgpack.Sample{Foo: "test-resp", Bar: req.Bar + 1})
	})
	go s.Serve(":0") // nolint
	defer s.Stop()   // nolint

	<-s.accepting

	// client
	client, err := net.Dial("tcp", s.Listener.Addr().String())
	if err != nil {
		panic(err)
	}
	defer client.Close() // nolint

	data, _ := s.Codec.Encode(&msgpack.Sample{Foo: "test", Bar: 1})
	packedMsg, _ := s.Packer.Pack(&message.Entry{ID: 1, Data: data})
	beforeBench(b)
	for i := 0; i < b.N; i++ {
		_, _ = client.Write(packedMsg)
	}
}

func beforeBench(b *testing.B) {
	Log = &mutedLogger{}
	b.ReportAllocs()
	b.ResetTimer()
}
