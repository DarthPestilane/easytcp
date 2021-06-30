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

	go func() {
		for {
			data := []byte("hello")
			msg := &packet.MessageEntry{
				ID:   1,
				Data: data,
			}
			msgSend, _ := packer.Pack(msg)
			if _, err := conn.Write(msgSend); err != nil {
				panic(err)
			}
			time.Sleep(time.Second)
		}
	}()

	for {
		buff := make([]byte, 1024)
		n, err := conn.Read(buff)
		if err != nil {
			panic(err)
		}
		msg, _ := packer.Unpack(bytes.NewReader(buff[:n]))
		fmt.Printf("recv: %s\n", msg.Data)
	}
}
