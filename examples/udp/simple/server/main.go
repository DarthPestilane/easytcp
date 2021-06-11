package main

import (
	"fmt"
	"github.com/DarthPestilane/easytcp"
	"github.com/DarthPestilane/easytcp/examples/fixture"
	"github.com/DarthPestilane/easytcp/logger"
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/DarthPestilane/easytcp/router"
	"github.com/DarthPestilane/easytcp/server"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"runtime"
	"time"
)

var log *logrus.Logger

func init() {
	log = logger.Default
}

func main() {
	// go printGoroutineNum()
	s := easytcp.NewUDPServer(server.UDPOption{})

	s.AddRoute(1, func(ctx *router.Context) (*packet.Response, error) {
		log.Infof("recv: %s", string(ctx.Request.RawData))
		return &packet.Response{ID: 2, Data: "done"}, nil
	})

	go func() {
		log.Infof("serve at %s", fixture.ServerAddr)
		if err := s.Serve(fixture.ServerAddr); err != nil {
			log.Errorf("serve err: %s", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	<-sigCh
	if err := s.Stop(); err != nil {
		log.Errorf("server stopped err: %s", err)
	}
	// <-time.After(time.Second * 5)
}

// nolint
func printGoroutineNum() {
	for {
		fmt.Println("goroutine num: ", runtime.NumGoroutine())
		time.Sleep(time.Second)
	}
}
