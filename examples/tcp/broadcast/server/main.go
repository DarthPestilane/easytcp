package main

import (
	"fmt"
	"github.com/DarthPestilane/easytcp"
	"github.com/DarthPestilane/easytcp/examples/fixture"
	"github.com/DarthPestilane/easytcp/examples/tcp/broadcast/common"
	"github.com/DarthPestilane/easytcp/message"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
}

func main() {
	s := easytcp.NewServer(&easytcp.ServerOption{
		Packer: &easytcp.DefaultPacker{},
	})

	s.Use(fixture.RecoverMiddleware(log), logMiddleware)

	s.AddRoute(common.MsgIdBroadCastReq, func(ctx *easytcp.Context) (*message.Entry, error) {
		var reqData string
		_ = ctx.Bind(&reqData)

		// broadcasting
		go easytcp.Sessions().Range(func(id string, sess *easytcp.Session) (next bool) {
			if ctx.Session().ID() == id {
				return true // next iteration
			}
			msg, err := ctx.Response(common.MsgIdBroadCastAck, fmt.Sprintf("%s (broadcast from %s)", reqData, ctx.Session().ID()))
			if err != nil {
				log.Errorf("create response err: %s", err)
				return true
			}
			if err := sess.SendResp(msg); err != nil {
				log.Errorf("broadcast err: %s", err)
			}
			return true
		})

		return ctx.Response(common.MsgIdBroadCastAck, "broadcast done")
	})

	go func() {
		if err := s.Serve(fixture.ServerAddr); err != nil {
			log.Error(err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	<-sigCh
	if err := s.Stop(); err != nil {
		log.Errorf("server stopped err: %s", err)
	}
}

func logMiddleware(next easytcp.HandlerFunc) easytcp.HandlerFunc {
	return func(ctx *easytcp.Context) (resp *message.Entry, err error) {
		log.Infof("recv request | %s", ctx.Message().Data)
		defer func() {
			if err != nil || resp == nil {
				return
			}
			r, _ := ctx.Get(easytcp.RespKey)
			log.Infof("send response | id: %d; size: %d; data: %s", resp.ID, len(resp.Data), r)
		}()
		return next(ctx)
	}
}
