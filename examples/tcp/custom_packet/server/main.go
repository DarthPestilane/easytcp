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
	log.SetLevel(logrus.DebugLevel)
}

func main() {
	easytcp.SetLogger(log)

	s := easytcp.NewTcp(server.TcpOption{
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

func handler(s session.Session, req *packet.Request) (*packet.Response, error) {
	var data fixture.Json01Req
	_ = s.MsgCodec().Decode(req.RawData, &data)

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

func logMiddleware(next router.HandlerFunc) router.HandlerFunc {
	return func(s session.Session, req *packet.Request) (resp *packet.Response, err error) {
		var data fixture.Json01Req
		_ = s.MsgCodec().Decode(req.RawData, &data)
		log.Infof("recv request | id:(%d) size:(%d) data: %+v", req.Id, req.RawSize, data)

		defer func() {
			if err == nil {
				size := 0
				if resp != nil {
					msgData, _ := s.MsgCodec().Encode(resp.Data)
					size = len(msgData)
					log.Infof("send response | id:(%d) size:(%d) data: %+v", resp.Id, size, resp.Data)
				} else {
					log.Infof("don't send response since nil")
				}
			}
		}()
		return next(s, req)
	}
}
