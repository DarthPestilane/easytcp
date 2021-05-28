package main

import (
	"fmt"
	"github.com/DarthPestilane/easytcp/tests/fixture"
	"net"
	"time"
)

func main() {
	conn, err := net.Dial("tcp", fixture.ServerAddr)
	if err != nil {
		panic(err)
	}
	codec := &fixture.JsonCodec{}
	packer := &fixture.Packer16bit{}
	go func() {
		// write loop
		for {
			time.Sleep(time.Second)
			req := map[string]interface{}{
				"bool":   true,
				"string": "string",
				"number": 123,
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
			data, err := codec.Decode(msg.GetData())
			if err != nil {
				panic(err)
			}
			fmt.Printf("recv ==> id:(%d) len:(%d) data: %+v\n", msg.GetId(), msg.GetSize(), data)
		}
	}()
	select {}
}
