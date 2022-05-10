package main

import (
	"github.com/DarthPestilane/easytcp"
	"github.com/DarthPestilane/easytcp/internal/examples/fixture"
	common2 "github.com/DarthPestilane/easytcp/internal/examples/tcp/proto_packet/common"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
	log.SetLevel(logrus.DebugLevel)
}

func main() {
	srv := easytcp.NewServer(&easytcp.ServerOption{
		Packer: &common2.CustomPacker{},
		Codec:  &easytcp.ProtobufCodec{},
	})

	srv.AddRoute(common2.ID_FooReqID, handle, logTransmission(&common2.FooReq{}, &common2.FooResp{}))

	if err := srv.Serve(fixture.ServerAddr); err != nil {
		log.Errorf("serve err: %s", err)
	}
}

func handle(c easytcp.Context) {
	var reqData common2.FooReq
	_ = c.Bind(&reqData)
	err := c.SetResponse(common2.ID_FooRespID, &common2.FooResp{
		Code:    2,
		Message: "success",
	})
	if err != nil {
		log.Errorf("set response failed: %s", err)
	}
}

func logTransmission(req, resp proto.Message) easytcp.MiddlewareFunc {
	return func(next easytcp.HandlerFunc) easytcp.HandlerFunc {
		return func(c easytcp.Context) {
			if err := c.Bind(req); err == nil {
				log.Debugf("recv | id: %d; size: %d; data: %s", c.Request().ID(), len(c.Request().Data()), req)
			}
			defer func() {
				respMsg := c.Response()
				if respMsg != nil {
					_ = c.Session().Codec().Decode(respMsg.Data(), resp)
					log.Infof("send | id: %d; size: %d; data: %s", respMsg.ID(), len(respMsg.Data()), resp)
				}
			}()
			next(c)
		}
	}
}
