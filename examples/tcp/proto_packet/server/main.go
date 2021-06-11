package main

import (
	"github.com/DarthPestilane/easytcp"
	"github.com/DarthPestilane/easytcp/examples/fixture"
	"github.com/DarthPestilane/easytcp/examples/tcp/proto_packet/message"
	"github.com/DarthPestilane/easytcp/logger"
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/DarthPestilane/easytcp/router"
	"github.com/DarthPestilane/easytcp/server"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log = logger.Default
}

func main() {
	srv := easytcp.NewTCPServer(server.TCPOption{
		MsgCodec: &fixture.ProtoCodec{},
	})

	srv.AddRoute(uint(message.ID_FooReqID), handle, logMiddleware)

	log.Infof("serve at %s", fixture.ServerAddr)
	if err := srv.Serve(fixture.ServerAddr); err != nil {
		log.Errorf("serve err: %s", err)
	}
}

func handle(ctx *router.Context) (*packet.Response, error) {
	var reqData message.FooReq
	if err := ctx.Bind(&reqData); err != nil {
		return nil, err
	}
	return &packet.Response{
		ID: uint(message.ID_FooRespID),
		Data: &message.FooResp{
			Code:    2,
			Message: "success",
		},
	}, nil
}

func logMiddleware(next router.HandlerFunc) router.HandlerFunc {
	return func(ctx *router.Context) (*packet.Response, error) {
		var reqData message.FooReq
		if err := ctx.Bind(&reqData); err == nil {
			log.Debugf("recv | id: %d; size: %d; data: %s", ctx.MessageID(), ctx.MessageSize(), reqData.String())
		}
		resp, err := next(ctx)
		if err != nil {
			return resp, err
		}
		if resp != nil {
			if msg, err := ctx.Session.MsgCodec().Encode(resp.Data); err == nil {
				log.Infof("send | id: %d; size: %d; data: %s", resp.ID, len(msg), resp.Data)
			}
		}
		return resp, err
	}
}
