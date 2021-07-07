package main

import (
	"fmt"
	"github.com/DarthPestilane/easytcp"
	"github.com/DarthPestilane/easytcp/examples/fixture"
	"github.com/DarthPestilane/easytcp/message"
	"github.com/sirupsen/logrus"
	"math/rand"
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
		// customize codec and packer
		Codec:  &easytcp.JsonCodec{},
		Packer: &fixture.Packer16bit{},
	})

	s.AddRoute(fixture.MsgIdJson01Req, handler, fixture.RecoverMiddleware(log), logMiddleware)

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

	// make a random panic
	if rand.Intn(2) == 0 {
		panic("random panic here")
	}

	return ctx.Response(fixture.MsgIdJson01Ack, &fixture.Json01Resp{
		Success: true,
		Data:    fmt.Sprintf("%s:%d:%t", data.Key1, data.Key2, data.Key3),
	})
}

func logMiddleware(next easytcp.HandlerFunc) easytcp.HandlerFunc {
	return func(ctx *easytcp.Context) (resp *message.Entry, err error) {
		// var data fixture.Json01Req
		// _ = ctx.Bind(&data)
		log.Infof("recv request | id:(%d) size:(%d) data: %s", ctx.MsgID(), ctx.MsgSize(), ctx.MsgData())

		defer func() {
			if err != nil {
				return
			}
			if resp != nil {
				log.Infof("send response | id:(%d) size:(%d) data: %s", resp.ID, len(resp.Data), resp.Data)
			} else {
				log.Infof("don't send response since nil")
			}
		}()
		return next(ctx)
	}
}
