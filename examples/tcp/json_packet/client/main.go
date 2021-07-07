package main

import (
	"github.com/DarthPestilane/easytcp"
	"github.com/DarthPestilane/easytcp/examples/fixture"
	"github.com/DarthPestilane/easytcp/message"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

func main() {
	conn, err := net.Dial("tcp", fixture.ServerAddr)
	if err != nil {
		panic(err)
	}
	log := logrus.New()
	codec := &easytcp.JsonCodec{}
	packer := &fixture.Packer16bit{}
	go func() {
		// write loop
		for {
			time.Sleep(time.Second)
			req := &fixture.Json01Req{
				Key1: "hello",
				Key2: 10,
				Key3: true,
			}
			data, err := codec.Encode(req)
			if err != nil {
				panic(err)
			}
			msg := &message.Entry{
				ID:   fixture.MsgIdJson01Req,
				Data: data,
			}
			packedMsg, err := packer.Pack(msg)
			if err != nil {
				panic(err)
			}
			if _, err := conn.Write(packedMsg); err != nil {
				panic(err)
			}
		}
	}()
	go func() {
		// read loop
		for {
			msg, err := packer.Unpack(conn)
			if err != nil {
				panic(err)
			}
			// var data fixture.Json01Resp
			// if err := codec.Decode(msg.Data, &data); err != nil {
			// 	panic(err)
			// }
			log.Infof("ack received | id:(%d) size:(%d) data: %s", msg.ID, len(msg.Data), msg.Data)
		}
	}()
	select {}
}
