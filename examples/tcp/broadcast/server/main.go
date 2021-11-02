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
	"time"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
}

func main() {
	s := easytcp.NewServer(&easytcp.ServerOption{
		Packer: easytcp.NewDefaultPacker(),
	})

	s.Use(fixture.RecoverMiddleware(log), logMiddleware)

	s.AddRoute(common.MsgIdBroadCastReq, func(ctx easytcp.Context) {
		reqData := ctx.Request().Data

		// broadcasting
		go easytcp.Sessions().Range(func(id string, sess easytcp.Session) (next bool) {
			if ctx.Session().ID() == id {
				return true // next iteration
			}
			respData := fmt.Sprintf("%s (broadcast from %s)", reqData, ctx.Session().ID())
			ctx.Copy().SetResponseMessage(&message.Entry{
				ID:   common.MsgIdBroadCastAck,
				Data: []byte(respData),
			}).SendTo(sess)
			return true
		})

		ctx.SetResponseMessage(&message.Entry{
			ID:   common.MsgIdBroadCastAck,
			Data: []byte("broadcast done"),
		})
	})

	go func() {
		if err := s.Serve(fixture.ServerAddr); err != nil {
			log.Error(err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	if err := s.Stop(); err != nil {
		log.Errorf("server stopped err: %s", err)
	}
	time.Sleep(time.Second)
}

func logMiddleware(next easytcp.HandlerFunc) easytcp.HandlerFunc {
	return func(ctx easytcp.Context) {
		log.Infof("recv request | %s", ctx.Request().Data)
		defer func() {
			var resp = ctx.Response()
			log.Infof("send response | id: %d; size: %d; data: %s", resp.ID, len(resp.Data), resp.Data)
		}()
		next(ctx)
	}
}
