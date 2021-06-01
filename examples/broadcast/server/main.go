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
)

var log *logrus.Logger

func init() {
	log = logger.Default
}

func main() {
	s := easytcp.NewTcp(server.TcpOption{})

	easytcp.RegisterMiddleware(logMiddleware)

	easytcp.RegisterRoute(fixture.MsgIdBroadCastReq, func(s *session.Session, req *packet.Request) (*packet.Response, error) {
		var reqData string
		_ = s.MsgCodec.Decode(req.RawData, &reqData)
		session.Sessions().Range(func(id string, sess *session.Session) (next bool) {
			if s.Id == id {
				return true // next iteration
			}
			err := sess.SendResp(&packet.Response{
				Id:   fixture.MsgIdBroadCastAck,
				Data: fmt.Sprintf("%s (broadcast from %s)", reqData, s.Id),
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

	if err := s.GracefulStop(); err != nil {
		log.Error(err)
	}
}

func logMiddleware(next router.HandlerFunc) router.HandlerFunc {
	return func(s *session.Session, req *packet.Request) (*packet.Response, error) {
		log.Infof("recv request | %s", req.RawData)
		return next(s, req)
	}
}
