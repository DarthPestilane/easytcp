package main

import (
	"fmt"
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/DarthPestilane/easytcp/tests/fixture"
	"net"
	"time"
)

func main() {
	conn, err := net.Dial("tcp", fixture.ServerAddr)
	if err != nil {
		panic(err)
	}
	codec := &packet.DefaultCodec{}
	packer := &packet.DefaultPacker{}
	go func() {
		// write loop
		for {
			time.Sleep(time.Second)
			data, err := codec.Encode("ping,ping,ping")
			if err != nil {
				panic(err)
			}
			msg, err := packer.Pack(fixture.MsgIdPingReq, data)
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
			data, err := codec.Decode(msg.GetData())
			if err != nil {
				panic(err)
			}
			fmt.Printf("recv ==> id:(%d) len:(%d) data: %s\n", msg.GetId(), msg.GetLen(), data)
		}
	}()
	select {}
}
