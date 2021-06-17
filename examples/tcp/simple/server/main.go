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
		MsgPacker:    &packet.DefaultPacker{}, // with default packer
		MsgCodec:     nil,                     // without codec
	})

	// register global middlewares
	s.Use(fixture.RecoverMiddleware(log), logMiddleware)

	// register a route
	s.AddRoute(fixture.MsgIdPingReq, func(ctx *router.Context) (packet.Message, error) {
		return ctx.Response(fixture.MsgIdPingAck, "pong, pong, pong")
	})

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

func logMiddleware(next router.HandlerFunc) router.HandlerFunc {
	return func(ctx *router.Context) (resp packet.Message, err error) {
		log.Infof("rec <<< | id:(%d) size:(%d) data: %s", ctx.MsgID(), ctx.MsgSize(), ctx.MsgRawData())
		defer func() {
			if err != nil || resp == nil {
				return
			}
			log.Infof("snd >>> | id:(%d) size:(%d) data: %s", resp.GetID(), resp.GetSize(), ctx.Value(router.RespKey))
		}()
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
