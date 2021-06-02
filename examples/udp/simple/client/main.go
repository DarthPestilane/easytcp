package main

import (
	"bytes"
	"fmt"
	"github.com/DarthPestilane/easytcp/examples/fixture"
	"github.com/DarthPestilane/easytcp/packet"
	"net"
	"time"
)

func main() {
	conn, err := net.Dial("udp", fixture.ServerAddr)
	if err != nil {
		panic(err)
	}
	packer := &packet.DefaultPacker{}

	time.Sleep(time.Second)
	msgSend, _ := packer.Pack(1, []byte("hello"))
	if _, err := conn.Write(msgSend); err != nil {
		panic(err)
	}

	buff := make([]byte, 1024)
	n, err := conn.Read(buff)
	if err != nil {
		panic(err)
	}
	msg, _ := packer.Unpack(bytes.NewReader(buff[:n]))
	fmt.Printf("recv: %s\n", msg.GetData())
}
