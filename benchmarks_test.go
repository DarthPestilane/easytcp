package easytcp

import (
	"github.com/DarthPestilane/easytcp/message"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

// go test -bench="^BenchmarkTCPServer_\w+$" -run=none -benchmem -benchtime=250000x

type mutedLogger struct{}

func (m *mutedLogger) Errorf(_ string, _ ...interface{}) {}
func (m *mutedLogger) Tracef(_ string, _ ...interface{}) {}

func Benchmark_NoHandler(b *testing.B) {
	s := NewServer(&ServerOption{
		DoNotPrintRoutes: true,
	})
	go s.Serve("127.0.0.1:0") // nolint
	defer s.Stop()            // nolint

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
	s.AddRoute(1, func(ctx Context) {})
	go s.Serve("127.0.0.1:0") // nolint
	defer s.Stop()            // nolint

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

func Benchmark_DefaultPacker_Pack(b *testing.B) {
	packer := NewDefaultPacker()

	msg := &message.Entry{
		ID:   1,
		Data: []byte("test"),
	}
	beforeBench(b)
	for i := 0; i < b.N; i++ {
		_, _ = packer.Pack(msg)
	}
}

func Benchmark_DefaultPacker_Unpack(b *testing.B) {
	packer := NewDefaultPacker()
	msg := &message.Entry{
		ID:   1,
		Data: []byte("test"),
	}
	bytes, err := packer.Pack(msg)
	assert.NoError(b, err)

	p1, p2 := net.Pipe()
	go func() {
		for {
			_, _ = p1.Write(bytes)
		}
	}()
	beforeBench(b)
	for i := 0; i < b.N; i++ {
		_, _ = packer.Unpack(p2)
	}
}

func beforeBench(b *testing.B) {
	Log = &mutedLogger{}
	b.ReportAllocs()
	b.ResetTimer()
}
