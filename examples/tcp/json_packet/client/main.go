package main

import (
	"github.com/DarthPestilane/easytcp/examples/fixture"
	"github.com/DarthPestilane/easytcp/logger"
	"net"
	"time"
)

func main() {
	conn, err := net.Dial("tcp", fixture.ServerAddr)
	if err != nil {
		panic(err)
	}
	log := logger.Default
	codec := &fixture.JsonCodec{}
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
			msg, err := packer.Pack(fixture.MsgIdJson01Req, data)
			if err != nil {
				panic(err)
			}
			if _, err := conn.Write(msg); err != nil {
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
			var data fixture.Json01Resp
			if err := codec.Decode(msg.GetData(), &data); err != nil {
				panic(err)
			}
			log.Infof("ack received | id:(%d) size:(%d) data: %+v", msg.GetID(), msg.GetSize(), data)
		}
	}()
	select {}
}
