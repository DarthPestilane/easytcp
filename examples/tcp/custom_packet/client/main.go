package main

import (
	"github.com/DarthPestilane/easytcp"
	"github.com/DarthPestilane/easytcp/examples/fixture"
	"github.com/DarthPestilane/easytcp/examples/tcp/custom_packet/common"
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
	packer := &common.CustomPacker{}
	go func() {
		// write loop
		for {
			time.Sleep(time.Second)
			req := &common.Json01Req{
				Key1: "hello",
				Key2: 10,
				Key3: true,
			}
			data, err := codec.Encode(req)
			if err != nil {
				panic(err)
			}
			msg := &message.Entry{
				ID:   "json01-req",
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
			fullSize := msg.MustGet("fullSize")
			log.Infof("ack received | fullSize:(%d) id:(%v) dataSize:(%d) data: %s", fullSize, msg.ID, len(msg.Data), msg.Data)
		}
	}()
	select {}
}
