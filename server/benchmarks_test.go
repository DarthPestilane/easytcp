package server

import (
	"github.com/DarthPestilane/easytcp/logger"
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/DarthPestilane/easytcp/router"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net"
	"testing"
)

//go:generate go test -bench=^BenchmarkTCP$ -run=none -benchmem -memprofile=bench_profiles/BenchmarkTCP.mem.out -cpuprofile=bench_profiles/BenchmarkTCP.profile.out
func BenchmarkTCP(b *testing.B) {
	muteLog()
	packer := &packet.DefaultPacker{}
	s := NewTCPServer(&TCPOption{
		MsgPacker: packer,
	})
	s.AddRoute(1, func(ctx *router.Context) (packet.Message, error) {
		return ctx.Response(2, "bench done")
	})
	go s.Serve(":0") // nolint
	defer s.Stop()   // nolint

	<-s.accepting

	// client
	client, err := net.Dial("tcp", s.listener.Addr().String())
	if err != nil {
		panic(err)
	}

	rawData := []byte("bench me")
	msg := &packet.DefaultMsg{
		ID:   1,
		Size: uint32(len(rawData)),
		Data: rawData,
	}
	packedMsg, err := packer.Pack(msg)
	if err != nil {
		panic(err)
	}
	benchRequest(b, client, packedMsg)
}

//go:generate go test -bench=^BenchmarkUDP$ -run=none -benchmem -memprofile=bench_profiles/BenchmarkUDP.mem.out -cpuprofile=bench_profiles/BenchmarkUDP.profile.out
func BenchmarkUDP(b *testing.B) {
	muteLog()
	packer := &packet.DefaultPacker{}
	s := NewUDPServer(&UDPOption{
		MaxBufferSize: 100,
		MsgPacker:     packer,
	})
	s.AddRoute(1, func(ctx *router.Context) (packet.Message, error) {
		return ctx.Response(2, "bench done")
	})
	go s.Serve(":0") // nolint
	defer s.Stop()   // nolint

	<-s.accepting
	client, err := net.Dial("udp", s.conn.LocalAddr().String())
	if err != nil {
		panic(err)
	}

	rawData := []byte("bench me")
	msg := &packet.DefaultMsg{
		ID:   1,
		Size: uint32(len(rawData)),
		Data: rawData,
	}
	packedMsg, err := packer.Pack(msg)
	if err != nil {
		panic(err)
	}
	benchRequest(b, client, packedMsg)
}

func benchRequest(b *testing.B, client net.Conn, msg []byte) {
	for i := 0; i < b.N; i++ {
		if _, err := client.Write(msg); err != nil {
			panic(err)
		}
	}
}

func muteLog() {
	log := logrus.New()
	log.SetOutput(ioutil.Discard)
	logger.Default = log
}
