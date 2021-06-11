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
	"syscall"
)

var log *logrus.Logger

func init() {
	log = logger.Default
	log.SetLevel(logrus.DebugLevel)
}

func main() {
	easytcp.SetLogger(log)

	s := easytcp.NewTCPServer(server.TCPOption{
		// customize codec and packer
		MsgCodec:  &fixture.JsonCodec{},
		MsgPacker: &fixture.Packer16bit{},
	})

	s.AddRoute(fixture.MsgIdJson01Req, handler, fixture.RecoverMiddleware(log), logMiddleware)

	go func() {
		log.Infof("serve at %s", fixture.ServerAddr)
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

func handler(ctx *router.Context) (packet.Message, error) {
	var data fixture.Json01Req
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

	return ctx.Response(fixture.MsgIdJson01Ack, &fixture.Json01Resp{
		Success: true,
		Data:    fmt.Sprintf("%s:%d:%t", data.Key1, data.Key2, data.Key3),
	})
}

func logMiddleware(next router.HandlerFunc) router.HandlerFunc {
	return func(ctx *router.Context) (resp packet.Message, err error) {
		var data fixture.Json01Req
		_ = ctx.Bind(&data)
		log.Infof("recv request | id:(%d) size:(%d) data: %+v", ctx.MsgID(), ctx.MsgSize(), data)

		defer func() {
			if err == nil {
				if resp != nil {
					log.Infof("send response | id:(%d) size:(%d) data: %s", resp.GetID(), resp.GetSize(), resp.GetData())
				} else {
					log.Infof("don't send response since nil")
				}
			}
		}()
		return next(ctx)
	}
}
