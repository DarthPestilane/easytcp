package main

import (
	"github.com/DarthPestilane/easytcp"
	"github.com/DarthPestilane/easytcp/codec"
	v1 "github.com/DarthPestilane/easytcp/example/proto/hello_world/v1"
	"github.com/sirupsen/logrus"
	"os"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)

	s := easytcp.NewServer("127.0.0.1", 7777)

	s.OnConnected(func(conn *easytcp.Connection) {
		logrus.Infof("connected! hello %s", conn.RemoteAddr())
		_ = conn.Send("", []byte("talk, now!"))
	})

	s.OnDisconnect(func(conn *easytcp.Connection) {
		logrus.Warnf("disconnect! bye bye %s", conn.RemoteAddr())
	})

	s.AddRoute("agent->backend", func(ctx *easytcp.Context) {
		var req v1.DemoAgent
		if err := ctx.Bind(codec.DefaultProtobuf, &req); err != nil {
			logrus.Errorf("bind error: %s", err)
			return
		}
		logrus.Infof("received from agent: %s", req.String())
		resp := &v1.DemoResp{Value: "nice! " + req.Proxy}
		b, err := codec.DefaultProtobuf.Marshal(resp)
		if err != nil {
			logrus.Error(err)
			return
		}
		if err := ctx.Conn().SendBuffer("backend->agent", b); err != nil {
			logrus.Errorf("write to agent failed: %s", err)
			return
		}
	})

	s.AddRoute("text", func(ctx *easytcp.Context) {
		logrus.Infof("recieved: %s", ctx.Body())
		_ = ctx.Conn().Send("text", []byte("copy that"))
	})

	logrus.Infof("backend server on : %s:%d", s.Addr, s.Port)
	if err := s.Serve(); err != nil {
		logrus.Errorf("serve failed: %s", err)
		os.Exit(1)
	}
}
