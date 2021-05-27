package main

import (
	"fmt"
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
	s := server.NewTcp(server.Option{
		RWBufferSize: 1024 * 1024,
	})

	router.Inst().Register(fixture.MsgIdPing, func(s *session.Session, req *packet.Request) *packet.Response {
		fmt.Println("request: ", req.Data)
		return &packet.Response{
			Id:   req.Id,
			Data: "pong,pong,pong",
		}
	})

	go func() {
		if err := s.Serve("127.0.0.1:8888"); err != nil {
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
