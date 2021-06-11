package main

import (
	"github.com/DarthPestilane/easytcp/examples/fixture"
	"github.com/DarthPestilane/easytcp/examples/tcp/proto_packet/message"
	"github.com/DarthPestilane/easytcp/logger"
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

var log *logrus.Logger

func init() {
	log = logger.Default
}

func main() {
	conn, err := net.Dial("tcp", fixture.ServerAddr)
	if err != nil {
		panic(err)
	}

	packer := &packet.DefaultPacker{}
	codec := &fixture.ProtoCodec{}

	go func() {
		for {
			var id = uint(message.ID_FooReqID)
			req := &message.FooReq{
				Bar: "bar",
				Buz: 22,
			}
			data, err := codec.Encode(req)
			if err != nil {
				panic(err)
			}
			msg := &packet.DefaultMsg{
				ID:   uint32(id),
				Size: uint32(len(data)),
				Data: data,
			}
			packedMsg, err := packer.Pack(msg)
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
		var respData message.FooResp
		if err := codec.Decode(msg.GetData(), &respData); err != nil {
			panic(err)
		}
		log.Infof("recv | id: %d; size: %d; data: %s", msg.GetID(), msg.GetSize(), respData.String())
	}
}
