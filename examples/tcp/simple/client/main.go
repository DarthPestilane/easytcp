package main

import (
	"github.com/DarthPestilane/easytcp/examples/fixture"
	"github.com/DarthPestilane/easytcp/logger"
	"github.com/DarthPestilane/easytcp/packet"
	"net"
	"time"
)

func main() {
	conn, err := net.Dial("tcp", fixture.ServerAddr)
	if err != nil {
		panic(err)
	}
	log := logger.Default
	packer := &packet.DefaultPacker{}
	go func() {
		// write loop
		for {
			time.Sleep(time.Second)
			rawData := []byte("ping, ping, ping")
			msg := &packet.DefaultMsg{
				ID:   uint32(fixture.MsgIdPingReq),
				Size: uint32(len(rawData)),
				Data: rawData,
			}
			packedMsg, err := packer.Pack(msg)
			if err != nil {
				panic(err)
			}
			if _, err := conn.Write(packedMsg); err != nil {
				panic(err)
			}
			log.Infof("snd >>> | id:(%d) size:(%d) data: %s", msg.GetID(), msg.GetSize(), rawData)
		}
	}()
	go func() {
		// read loop
		for {
			msg, err := packer.Unpack(conn)
			if err != nil {
				panic(err)
			}
			log.Infof("rec <<< | id:(%d) size:(%d) data: %s", msg.GetID(), msg.GetSize(), msg.GetData())
		}
	}()
	select {}
}
