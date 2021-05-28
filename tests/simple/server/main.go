package main

import (
	"fmt"
	"github.com/DarthPestilane/easytcp"
	"github.com/DarthPestilane/easytcp/logger"
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/DarthPestilane/easytcp/server"
	"github.com/DarthPestilane/easytcp/session"
	"github.com/DarthPestilane/easytcp/tests/fixture"
	"runtime"
	"time"
)

func main() {
	// go printGoroutineNum()

	log := logger.Default

	s := easytcp.NewTcp(server.TcpOption{
		RWBufferSize: 1024 * 1024,
	})

	easytcp.RegisterRoute(fixture.MsgIdPingReq, func(s *session.Session, req *packet.Request) *packet.Response {
		var data string
		_ = s.MsgCodec.Decode(req.RawData, &data)
		log.Debugf("request | id:(%d) size:(%d) data: %s", req.Id, req.RawSize, data)
		return &packet.Response{
			Id:   fixture.MsgIdPingAck,
			Data: "pong, pong, pong",
		}
	})

	go func() {
		if err := s.Serve(fixture.ServerAddr); err != nil {
			log.Errorf("serve err: %s", err)
		}
	}()

	if err := s.GracefulStop(); err != nil {
		panic(err)
	}

	time.Sleep(time.Second * 3)
}

// nolint: deadcode, unused
func printGoroutineNum() {
	for {
		fmt.Println("goroutine num: ", runtime.NumGoroutine())
		time.Sleep(time.Second)
	}
}
