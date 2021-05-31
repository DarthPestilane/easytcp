package main

import (
	"fmt"
	"github.com/DarthPestilane/easytcp"
	"github.com/DarthPestilane/easytcp/logger"
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/DarthPestilane/easytcp/router"
	"github.com/DarthPestilane/easytcp/server"
	"github.com/DarthPestilane/easytcp/session"
	"github.com/DarthPestilane/easytcp/tests/fixture"
	"github.com/sirupsen/logrus"
	"time"
)

var log *logrus.Logger

func init() {
	log = logger.Default
	log.SetLevel(logrus.DebugLevel)
}

func main() {
	easytcp.SetLogger(log)

	s := easytcp.NewTcp(server.TcpOption{
		// customize codec and packer
		MsgCodec:  &fixture.JsonCodec{},
		MsgPacker: &fixture.Packer16bit{},
	})

	easytcp.RegisterRoute(fixture.MsgIdJson01Req, handler, recoverMiddleware, logMiddleware)

	go func() {
		if err := s.Serve(fixture.ServerAddr); err != nil {
			log.Errorf("serve err: %s", err)
		}
	}()

	if err := s.GracefulStop(); err != nil {
		panic(err)
	}

	time.Sleep(time.Second * 3)
}

func handler(s *session.Session, req *packet.Request) (*packet.Response, error) {
	var data fixture.Json01Req
	_ = s.MsgCodec.Decode(req.RawData, &data)

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

	return &packet.Response{
		Id: fixture.MsgIdJson01Ack,
		Data: &fixture.Json01Resp{
			Success: true,
			Data:    fmt.Sprintf("%s:%d:%t", data.Key1, data.Key2, data.Key3),
		},
	}, nil
}

func recoverMiddleware(next router.HandlerFunc) router.HandlerFunc {
	return func(s *session.Session, req *packet.Request) (*packet.Response, error) {
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("PANIC | %+v", r)
			}
		}()
		return next(s, req)
	}
}

func logMiddleware(next router.HandlerFunc) router.HandlerFunc {
	return func(s *session.Session, req *packet.Request) (*packet.Response, error) {
		var data fixture.Json01Req
		_ = s.MsgCodec.Decode(req.RawData, &data)
		log.Infof("recv req | id:(%d) size:(%d) data: %+v", req.Id, req.RawSize, data)
		return next(s, req)
	}
}
