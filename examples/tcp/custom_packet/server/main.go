package main

import (
	"fmt"
	"github.com/DarthPestilane/easytcp"
	"github.com/DarthPestilane/easytcp/examples/fixture"
	"github.com/DarthPestilane/easytcp/message"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
	log.SetLevel(logrus.DebugLevel)
}

func main() {
	easytcp.Log = log

	s := easytcp.NewServer(&easytcp.ServerOption{
		// specify codec and packer
		Codec:  &easytcp.JsonCodec{},
		Packer: &fixture.CustomPacker{},
	})

	s.AddRoute("json01-req", handler, fixture.RecoverMiddleware(log), logMiddleware)

	go func() {
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

func handler(ctx *easytcp.Context) (*message.Entry, error) {
	var data fixture.Json01Req
	_ = ctx.Bind(&data)

	// make a random panic to exam the `fixture.RecoverMiddleware`
	// if rand.Intn(2) == 0 {
	// 	panic("random panic here")
	// }

	return ctx.Response("json01-resp", &fixture.Json01Resp{
		Success: true,
		Data:    fmt.Sprintf("%s:%d:%t", data.Key1, data.Key2, data.Key3),
	})
}

func logMiddleware(next easytcp.HandlerFunc) easytcp.HandlerFunc {
	return func(ctx *easytcp.Context) (resp *message.Entry, err error) {
		// var data fixture.Json01Req
		// _ = ctx.Bind(&data)
		fullSize, _ := ctx.Message().Get("fullSize")
		log.Infof("recv request  | fullSize:(%d) id:(%v) dataSize(%d) data: %s", fullSize, ctx.Message().ID, len(ctx.Message().Data), ctx.Message().Data)

		defer func() {
			if err != nil {
				return
			}
			if resp != nil {
				log.Infof("send response | dataSize:(%d) id:(%v) data: %s", len(resp.Data), resp.ID, resp.Data)
			} else {
				log.Infof("don't send response since nil")
			}
		}()
		return next(ctx)
	}
}
