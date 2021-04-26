package main

import (
	"demo/tcp_demo"
	"demo/tcp_demo/codec"
	v1 "demo/tcp_demo/proto/hello_world/v1"
	"github.com/sirupsen/logrus"
	"time"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func main() {
	clientToBackend := tcp_demo.NewClient("127.0.0.1", 7777)
	if err := clientToBackend.Dial(time.Second); err != nil {
		logrus.Errorf("dial backend failed: %s", err)
		return
	}

	agentChan := make(chan *v1.DemoResp)

	clientToBackend.AddRoute("backend->agent", func(ctx *tcp_demo.Context) {
		var resp v1.DemoResp
		if err := ctx.Bind(codec.DefaultProtobuf, &resp); err != nil {
			logrus.Errorf("bind error: %s", err)
			return
		}
		agentChan <- &resp
	})

	go clientToBackend.StartReading()

	agentSrv := tcp_demo.NewServer("127.0.0.1", 8888)

	agentSrv.AddRoute("client->agent", func(ctx *tcp_demo.Context) {
		var req v1.DemoReq
		if err := ctx.Bind(codec.DefaultProtobuf, &req); err != nil {
			logrus.Error(err)
			return
		}

		reqToSend := &v1.DemoAgent{Proxy: "key: " + req.Key}
		reqToSendByte, _ := codec.DefaultProtobuf.Marshal(reqToSend)

		// send to "agent->backend"
		_, _ = clientToBackend.SendIn("agent->backend", reqToSendByte, time.Second)

		// wait for response
		resp := <-agentChan
		logrus.Infof("received response from backend: %s", resp.String())

		b, err := codec.DefaultProtobuf.Marshal(resp)
		if err != nil {
			logrus.Error(err)
			return
		}

		// send back to client
		n, err := ctx.SendIn("agent->client", b, time.Second)
		if err != nil {
			logrus.Errorf("write to client failed: %s", err)
			return
		}
		logrus.Infof("write %d bytes to client", n)
	})

	logrus.Infof("agent server on: %s:%d", agentSrv.Addr, agentSrv.Port)
	if err := agentSrv.Serve(); err != nil {
		panic(err)
	}
}
