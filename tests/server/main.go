package main

import (
	"fmt"
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/DarthPestilane/easytcp/router"
	"github.com/DarthPestilane/easytcp/server"
	"github.com/DarthPestilane/easytcp/session"
	"runtime"
	"time"
)

const (
	_ uint32 = iota
	MsgIdPing
)

func main() {
	go printGoroutineNum()
	s := server.NewTcp(server.Option{
		RWBufferSize: 1024 * 1024,
	})

	router.Inst().Register(MsgIdPing, func(s *session.Session, msg *packet.Request) {
		fmt.Println("final msg: ", msg.Data.(string))
		s.Send([]byte("copy that"))
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

func printGoroutineNum() {
	for {
		fmt.Println("goroutine num: ", runtime.NumGoroutine())
		time.Sleep(time.Second)
	}
}
