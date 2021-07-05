package main

import (
	"fmt"
	"github.com/DarthPestilane/easytcp"
	"github.com/DarthPestilane/easytcp/examples/fixture"
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/DarthPestilane/easytcp/router"
	"github.com/DarthPestilane/easytcp/server"
	"github.com/DarthPestilane/easytcp/session"
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
	s := easytcp.NewTCPServer(&server.TCPOption{
		Packer: &packet.DefaultPacker{},
	})

	s.Use(fixture.RecoverMiddleware(log), logMiddleware)

	s.AddRoute(fixture.MsgIdBroadCastReq, func(ctx *router.Context) (*packet.MessageEntry, error) {
		var reqData string
		_ = ctx.Bind(&reqData)

		// broadcasting
		go session.Sessions().Range(func(id string, sess session.Session) (next bool) {
			if _, ok := sess.(*session.TCPSession); !ok { // only broadcast to the same kind sessions
				return true // next iteration
			}
			if ctx.SessionID() == id {
				return true // next iteration
			}
			msg, err := ctx.Response(fixture.MsgIdBroadCastAck, fmt.Sprintf("%s (broadcast from %s)", reqData, ctx.SessionID()))
			if err != nil {
				log.Errorf("create response err: %s", err)
				return true
			}
			if err := sess.SendResp(msg); err != nil {
				log.Errorf("broadcast err: %s", err)
			}
			return true
		})

		return ctx.Response(fixture.MsgIdBroadCastAck, "broadcast done")
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

func logMiddleware(next router.HandlerFunc) router.HandlerFunc {
	return func(ctx *router.Context) (resp *packet.MessageEntry, err error) {
		log.Infof("recv request | %s", ctx.MsgData())
		defer func() {
			if err != nil || resp == nil {
				return
			}
			r, _ := ctx.Get(router.RespKey)
			log.Infof("send response | id: %d; size: %d; data: %s", resp.ID, len(resp.Data), r)
		}()
		return next(ctx)
	}
}
