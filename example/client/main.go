package main

import (
	"github.com/DarthPestilane/easytcp"
	"github.com/DarthPestilane/easytcp/codec"
	v1 "github.com/DarthPestilane/easytcp/example/proto/hello_world/v1"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

const serverAddr = "127.0.0.1:7777"

func main() {
	netConn, err := net.DialTimeout("tcp", serverAddr, time.Second*5)
	if err != nil {
		logrus.Errorf("client dial failed: %s", err)
		return
	}
	logrus.Infof("client dial success")
	conn := easytcp.NewConnection(netConn, easytcp.ConnectionOption{BufferSize: 512})
	go conn.KeepWriting()
	for {
		time.Sleep(time.Second)
		req := &v1.DemoReq{Key: "demo"}
		data, err := codec.DefaultProtobuf.Marshal(req)
		if err != nil {
			logrus.Errorf("client codec marshal failed: %s", err)
			return
		}
		if err := conn.Send("protobuf", data); err != nil {
			logrus.Errorf("client send failed: %s", err)
			return
		}
		head, body, err := conn.ReadMessage()
		if err != nil {
			logrus.Errorf("client read conn failed: %s", err)
			return
		}
		var resp v1.DemoResp
		_ = codec.DefaultProtobuf.Unmarshal(body, &resp)
		logrus.Infof("received from server: [%s]%d|%s", head.RoutePath, head.Length, resp.String())
	}
}
