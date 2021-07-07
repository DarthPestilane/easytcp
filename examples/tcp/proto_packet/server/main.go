package main

import (
	"github.com/DarthPestilane/easytcp"
	"github.com/DarthPestilane/easytcp/examples/fixture"
	"github.com/DarthPestilane/easytcp/examples/tcp/proto_packet/message"
	message2 "github.com/DarthPestilane/easytcp/message"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
	log.SetLevel(logrus.DebugLevel)
}

func main() {
	srv := easytcp.NewServer(&easytcp.ServerOption{
		Packer: &easytcp.DefaultPacker{},
		Codec:  &fixture.ProtoCodec{},
	})

	srv.AddRoute(uint(message.ID_FooReqID), handle, logMiddleware)

	if err := srv.Serve(fixture.ServerAddr); err != nil {
		log.Errorf("serve err: %s", err)
	}
}

func handle(c *easytcp.Context) (*message2.Entry, error) {
	var reqData message.FooReq
	if err := c.Bind(&reqData); err != nil {
		return nil, err
	}
	return c.Response(uint(message.ID_FooRespID), &message.FooResp{
		Code:    2,
		Message: "success",
	})
}

func logMiddleware(next easytcp.HandlerFunc) easytcp.HandlerFunc {
	return func(c *easytcp.Context) (*message2.Entry, error) {
		var reqData message.FooReq
		if err := c.Bind(&reqData); err == nil {
			log.Debugf("recv | id: %d; size: %d; data: %s", c.Message().ID, len(c.Message().Data), reqData.String())
		}
		resp, err := next(c)
		if err != nil {
			return resp, err
		}
		if resp != nil {
			r, _ := c.Get(easytcp.RespKey)
			log.Infof("send | id: %d; size: %d; data: %s", resp.ID, len(resp.Data), r)
		}
		return resp, err
	}
}
