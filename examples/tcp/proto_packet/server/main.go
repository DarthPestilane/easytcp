package main

import (
	"github.com/DarthPestilane/easytcp"
	"github.com/DarthPestilane/easytcp/examples/fixture"
	"github.com/DarthPestilane/easytcp/examples/tcp/proto_packet/message"
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/DarthPestilane/easytcp/router"
	"github.com/DarthPestilane/easytcp/server"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
	log.SetLevel(logrus.DebugLevel)
}

func main() {
	srv := easytcp.NewTCPServer(&server.TCPOption{
		Packer: &packet.DefaultPacker{},
		Codec:  &fixture.ProtoCodec{},
	})

	srv.AddRoute(uint(message.ID_FooReqID), handle, logMiddleware)

	if err := srv.Serve(fixture.ServerAddr); err != nil {
		log.Errorf("serve err: %s", err)
	}
}

func handle(ctx *router.Context) (*packet.MessageEntry, error) {
	var reqData message.FooReq
	if err := ctx.Bind(&reqData); err != nil {
		return nil, err
	}
	return ctx.Response(uint(message.ID_FooReqID), &message.FooResp{
		Code:    2,
		Message: "success",
	})
}

func logMiddleware(next router.HandlerFunc) router.HandlerFunc {
	return func(ctx *router.Context) (*packet.MessageEntry, error) {
		var reqData message.FooReq
		if err := ctx.Bind(&reqData); err == nil {
			log.Debugf("recv | id: %d; size: %d; data: %s", ctx.MsgID(), ctx.MsgSize(), reqData.String())
		}
		resp, err := next(ctx)
		if err != nil {
			return resp, err
		}
		if resp != nil {
			r, _ := ctx.Get(router.RespKey)
			log.Infof("send | id: %d; size: %d; data: %s", resp.ID, len(resp.Data), r)
		}
		return resp, err
	}
}
