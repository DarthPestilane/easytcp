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
	s := easytcp.NewTcpServer(server.TcpOption{})

	s.Use(fixture.RecoverMiddleware(log), logMiddleware)

	s.AddRoute(fixture.MsgIdBroadCastReq, func(s session.Session, req *packet.Request) (*packet.Response, error) {
		var reqData string
		_ = s.MsgCodec().Decode(req.RawData, &reqData)
		session.Sessions().Range(func(id string, sess session.Session) (next bool) {
			if _, ok := sess.(*session.TCPSession); !ok {
				// only broadcast to the same kind sessions
				return true // next iteration
			}
			if s.ID() == id {
				return true // next iteration
			}
			_, err := sess.SendResp(&packet.Response{
				Id:   fixture.MsgIdBroadCastAck,
				Data: fmt.Sprintf("%s (broadcast from %s)", reqData, s.ID()),
			})
			if err != nil {
				log.Errorf("broadcast err: %s", err)
			}
			return true
		})
		return &packet.Response{Id: fixture.MsgIdBroadCastAck, Data: "broadcast done"}, nil
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
	return func(s session.Session, req *packet.Request) (*packet.Response, error) {
		log.Infof("recv request | %s", req.RawData)
		return next(s, req)
	}
}
