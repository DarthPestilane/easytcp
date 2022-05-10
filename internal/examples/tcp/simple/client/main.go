package main

import (
	"github.com/DarthPestilane/easytcp"
	"github.com/DarthPestilane/easytcp/internal/examples/fixture"
	"github.com/DarthPestilane/easytcp/internal/examples/tcp/simple/common"
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
	packer := easytcp.NewDefaultPacker()
	go func() {
		// write loop
		for {
			time.Sleep(time.Second)
			msg := easytcp.NewMessage(common.MsgIdPingReq, []byte("ping, ping, ping"))
			packedBytes, err := packer.Pack(msg)
			if err != nil {
				panic(err)
			}
			if _, err := conn.Write(packedBytes); err != nil {
				panic(err)
			}
			log.Infof("snd >>> | id:(%d) size:(%d) data: %s", msg.ID(), len(msg.Data()), msg.Data())
		}
	}()
	go func() {
		// read loop
		for {
			msg, err := packer.Unpack(conn)
			if err != nil {
				panic(err)
			}
			log.Infof("rec <<< | id:(%d) size:(%d) data: %s", msg.ID(), len(msg.Data()), msg.Data())
		}
	}()
	select {}
}
