package main

import (
	"github.com/DarthPestilane/easytcp"
	"github.com/DarthPestilane/easytcp/codec"
	"github.com/DarthPestilane/easytcp/core"
	v1 "github.com/DarthPestilane/easytcp/example/proto/hello_world/v1"
	"github.com/sirupsen/logrus"
	"os"
)

func main() {
	// set logger here
	// easytcp.SetLogger(...)

	s := easytcp.NewServer("127.0.0.1", 7777)

	s.SetBufferSize(512)

	s.OnConnected(func(conn *core.Connection) {
		logrus.Infof("connected! hello %s", conn.NetConn().RemoteAddr())
		// _ = conn.Send("", []byte("talk, now!"))
	})

	s.OnDisconnect(func(conn *core.Connection) {
		logrus.Warnf("disconnect! bye bye %s", conn.NetConn().RemoteAddr())
	})

	s.AddRoute("protobuf", func(ctx *core.Context) {
		var req v1.DemoAgent
		if err := ctx.Bind(codec.DefaultProtobuf, &req); err != nil {
			logrus.Errorf("bind error: %s", err)
			return
		}
		logrus.Infof("received from client: %s", req.String())
		resp := &v1.DemoResp{Value: "nice! " + req.Proxy}
		b, err := codec.DefaultProtobuf.Marshal(resp)
		if err != nil {
			logrus.Error(err)
			return
		}
		if err := ctx.Conn().Send("protobuf->client", b); err != nil {
			logrus.Errorf("write to agent failed: %s", err)
			return
		}
	})

	// telnet 127.0.0.2:7777>[text]10|hello world
	s.AddRoute("text", func(ctx *core.Context) {
		logrus.Infof("recieved: %s", ctx.Body())
		_ = ctx.Conn().Send("text", []byte("copy that"))
	})

	logrus.Infof("backend server on : %s:%d", s.Addr, s.Port)
	if err := s.Serve(); err != nil {
		logrus.Errorf("serve failed: %s", err)
		os.Exit(1)
	}
}
