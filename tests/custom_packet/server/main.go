package main

import (
	"github.com/DarthPestilane/easytcp"
	"github.com/DarthPestilane/easytcp/logger"
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/DarthPestilane/easytcp/server"
	"github.com/DarthPestilane/easytcp/session"
	"github.com/DarthPestilane/easytcp/tests/fixture"
	"github.com/sirupsen/logrus"
	"time"
)

func main() {
	log := logger.Default
	log.SetLevel(logrus.DebugLevel)
	easytcp.SetLogger(log)

	s := easytcp.NewTcp(server.TcpOption{
		// customize codec and packer
		MsgCodec:  &fixture.JsonCodec{},
		MsgPacker: &fixture.Packer16bit{},
	})

	easytcp.RegisterRoute(fixture.MsgIdJson01Req, func(s *session.Session, req *packet.Request) *packet.Response {
		log.Debugf("request ==> id:(%d) size:(%d) data: %+v", req.Id, req.RawSize, req.Data)

		resp := map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"key": "value",
			},
		}
		return &packet.Response{
			Id:   fixture.MsgIdJson01Ack,
			Data: resp,
		}
	})

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
