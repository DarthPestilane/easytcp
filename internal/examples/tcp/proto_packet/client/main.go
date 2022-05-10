package main

import (
	"github.com/DarthPestilane/easytcp"
	"github.com/DarthPestilane/easytcp/internal/examples/fixture"
	common2 "github.com/DarthPestilane/easytcp/internal/examples/tcp/proto_packet/common"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
	log.SetLevel(logrus.DebugLevel)
}

func main() {
	conn, err := net.Dial("tcp", fixture.ServerAddr)
	if err != nil {
		panic(err)
	}

	packer := &common2.CustomPacker{}
	codec := &easytcp.ProtobufCodec{}

	go func() {
		for {
			var id = common2.ID_FooReqID
			req := &common2.FooReq{
				Bar: "bar",
				Buz: 22,
			}
			data, err := codec.Encode(req)
			if err != nil {
				panic(err)
			}
			packedMsg, err := packer.Pack(easytcp.NewMessage(id, data))
			if err != nil {
				panic(err)
			}
			if _, err := conn.Write(packedMsg); err != nil {
				panic(err)
			}
			log.Debugf("send | id: %d; size: %d; data: %s", id, len(data), req.String())
			time.Sleep(time.Second)
		}
	}()

	for {
		msg, err := packer.Unpack(conn)
		if err != nil {
			panic(err)
		}
		var respData common2.FooResp
		if err := codec.Decode(msg.Data(), &respData); err != nil {
			panic(err)
		}
		log.Infof("recv | id: %d; size: %d; data: %s", msg.ID(), len(msg.Data()), respData.String())
	}
}
