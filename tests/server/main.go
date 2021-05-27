package main

import (
	"fmt"
	"github.com/DarthPestilane/easytcp/router"
	"github.com/DarthPestilane/easytcp/server"
	"github.com/DarthPestilane/easytcp/session"
	"runtime"
	"time"
)

func main() {
	go printGoroutineNum()
	s := server.NewTcp(server.Option{
		RWBufferSize: 1024 * 1024,
	})

	router.Inst().Register(uint32(1), func(s *session.Session, msg []byte) {
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
