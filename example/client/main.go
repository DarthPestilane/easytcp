package main

import (
	"demo/tcp_demo"
	"demo/tcp_demo/codec"
	v1 "demo/tcp_demo/proto/hello_world/v1"
	"demo/tcp_demo/util"
	"github.com/sirupsen/logrus"
	"time"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func main() {
	client := tcp_demo.NewClient("127.0.0.1", 8888)
	if err := client.Dial(time.Second); err != nil {
		logrus.Errorf("dial agent failed: %s", err)
		return
	}
	client.AddRoute("agent->client", func(ctx *tcp_demo.Context) {
		var resp v1.DemoResp
		if err := ctx.Bind(codec.DefaultProtobuf, &resp); err != nil {
			logrus.Errorf("bind error: %s", err)
			return
		}
		logrus.Infof("received response from agent: %s", resp.String())
	})

	go client.StartReading()

	for {
		time.Sleep(time.Second)

		req := &v1.DemoReq{Key: "client"}
		b, err := codec.DefaultProtobuf.Marshal(req)
		if err != nil {
			logrus.Error(err)
			return
		}
		n, err := client.SendIn("client->agent", b, time.Second)
		if err != nil {
			if util.IsEOF(err) {
				logrus.Errorf("agent disconnected!! %s", err)
				return
			}
			logrus.Errorf("write to agent failed: %s", err)
			continue
		}
		logrus.Infof("write %d bytes to agent", n)
	}
}
