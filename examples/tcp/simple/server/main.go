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
	"syscall"
	"time"
)

var log *logrus.Logger

func init() {
	log = logger.Default
}

func main() {
	// go printGoroutineNum()

	s := easytcp.NewTCPServer(server.TCPOption{
		RWBufferSize: 1024 * 1024,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	})

	s.AddRoute(fixture.MsgIdPingReq, handler, fixture.RecoverMiddleware(log), logMiddleware)

	go func() {
		log.Infof("serving at %s", fixture.ServerAddr)
		if err := s.Serve(fixture.ServerAddr); err != nil {
			log.Errorf("serve err: %s", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	<-sigCh
	if err := s.Stop(); err != nil {
		log.Errorf("server stopped err: %s", err)
	}
}

func handler(ctx *router.Context) (*packet.Response, error) {
	var data string
	_ = ctx.Bind(&data)

	panicMaker := map[bool]struct{}{
		true:  {},
		false: {},
	}
	for k := range panicMaker {
		if !k {
			panic("random panic here")
		}
		break
	}

	return &packet.Response{
		ID:   fixture.MsgIdPingAck,
		Data: data + "||pong, pong, pong",
	}, nil
}

func logMiddleware(next router.HandlerFunc) router.HandlerFunc {
	return func(ctx *router.Context) (*packet.Response, error) {
		var data string
		_ = ctx.Bind(&data)
		log.Infof("recv req | id:(%d) size:(%d) data: %s", ctx.MessageID(), ctx.MessageSize(), data)
		return next(ctx)
	}
}

// nolint: deadcode, unused
func printGoroutineNum() {
	for {
		fmt.Println("goroutine num: ", runtime.NumGoroutine())
		time.Sleep(time.Second)
	}
}
