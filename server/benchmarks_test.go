package server

import (
	"github.com/DarthPestilane/easytcp/logger"
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/DarthPestilane/easytcp/router"
	"net"
	"testing"
)

//go:generate go test -bench=^BenchmarkTCP$ -run=none -benchmem -memprofile=bench_profiles/BenchmarkTCP.mem.out -cpuprofile=bench_profiles/BenchmarkTCP.profile.out
func BenchmarkTCP(b *testing.B) {
	logger.Log = &logger.MuteLogger{}
	packer := &packet.DefaultPacker{}
	s := NewTCPServer(&TCPOption{
		MsgPacker:       packer,
		DontPrintRoutes: true,
	})
	s.AddRoute(1, func(ctx *router.Context) (*packet.MessageEntry, error) {
		return ctx.Response(2, []byte("pong"))
	})
	go s.Serve(":0") // nolint

	<-s.accepting

	// client
	client, err := net.Dial("tcp", s.listener.Addr().String())
	if err != nil {
		panic(err)
	}
	packedMsg, _ := packer.Pack(&packet.MessageEntry{ID: 1, Data: []byte("ping")})
	for i := 0; i < b.N; i++ {
		_, _ = client.Write(packedMsg)
	}
}
