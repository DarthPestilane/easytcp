package main

import (
	"fmt"
	"github.com/DarthPestilane/easytcp"
	"github.com/DarthPestilane/easytcp/examples/fixture"
	"github.com/DarthPestilane/easytcp/logger"
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
	log = logger.Default
}

func main() {
	s := easytcp.NewTCPServer(server.TCPOption{})

	s.Use(fixture.RecoverMiddleware(log), logMiddleware)

	s.AddRoute(fixture.MsgIdBroadCastReq, func(ctx *router.Context) (*packet.Response, error) {
		var reqData string
		_ = ctx.Bind(&reqData)
		session.Sessions().Range(func(id string, sess session.Session) (next bool) {
			if _, ok := sess.(*session.TCPSession); !ok {
				// only broadcast to the same kind sessions
				return true // next iteration
			}
			if ctx.Session.ID() == id {
				return true // next iteration
			}
			_, err := sess.SendResp(&packet.Response{
				ID:   fixture.MsgIdBroadCastAck,
				Data: fmt.Sprintf("%s (broadcast from %s)", reqData, ctx.Session.ID()),
			})
			if err != nil {
				log.Errorf("broadcast err: %s", err)
			}
			return true
		})
		return &packet.Response{ID: fixture.MsgIdBroadCastAck, Data: "broadcast done"}, nil
	})

	go func() {
		log.Infof("serve at %s", fixture.ServerAddr)
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
	return func(ctx *router.Context) (*packet.Response, error) {
		log.Infof("recv request | %s", ctx.MessageRawData())
		return next(ctx)
	}
}
