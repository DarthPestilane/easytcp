package main

import (
	"demo/tcp_demo"
	"demo/tcp_demo/codec"
	v1 "demo/tcp_demo/proto/hello_world/v1"
	"github.com/sirupsen/logrus"
	"time"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)

	s := tcp_demo.NewServer("127.0.0.1", 7777)

	s.AddRoute("agent->backend", func(ctx *tcp_demo.Context) {
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
		n, err := ctx.SendIn("backend->agent", b, time.Second)
		if err != nil {
			logrus.Errorf("write to agent failed: %s", err)
			return
		}
		logrus.Infof("write %d bytes to agent", n)
	})

	logrus.Infof("backend server on : %s:%d", s.Addr, s.Port)
	if err := s.Serve(); err != nil {
		panic(err)
	}
}
