package main

import (
	"fmt"
	"github.com/DarthPestilane/easytcp"
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/DarthPestilane/easytcp/router"
	"github.com/DarthPestilane/easytcp/server"
	"github.com/DarthPestilane/easytcp/session"
	"github.com/DarthPestilane/easytcp/tests/fixture"
	"runtime"
	"time"
)

func main() {
	// go printGoroutineNum()
	s := easytcp.NewTcp(server.TcpOption{
		RWBufferSize: 1024 * 1024,
	})

	router.Inst().Register(fixture.MsgIdPingReq, func(s *session.Session, req *packet.Request) *packet.Response {
		fmt.Printf("request ==> id:(%d) data: %s\n", req.Id, req.Data)
		return &packet.Response{
			Id:   fixture.MsgIdPingAck,
			Data: "pong,pong,pong",
		}
	})

	go func() {
		if err := s.Serve(fixture.ServerAddr); err != nil {
			panic(err)
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
