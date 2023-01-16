package main

import (
	"fmt"
	"github.com/DarthPestilane/easytcp"
	"github.com/DarthPestilane/easytcp/internal/examples/fixture"
	"github.com/DarthPestilane/easytcp/internal/examples/tcp/broadcast/common"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

var log *logrus.Logger
var packer easytcp.Packer

func init() {
	log = logrus.New()
	log.SetLevel(logrus.DebugLevel)
	packer = easytcp.NewDefaultPacker()
}

func main() {
	senderClient()
	for i := 0; i < 10; i++ {
		readerClient(i)
	}

	select {}
}

func establish() (net.Conn, error) {
	return net.Dial("tcp", fixture.ServerAddr)
}

func senderClient() {
	conn, err := establish()
	if err != nil {
		log.Error(err)
		return
	}
	// send
	go func() {
		for {
			time.Sleep(time.Second)
			data := []byte(fmt.Sprintf("hello everyone @%d", time.Now().Unix()))
			packedMsg, _ := packer.Pack(easytcp.NewMessage(common.MsgIdBroadCastReq, data))
			if _, err := conn.Write(packedMsg); err != nil {
				log.Error(err)
				return
			}
		}
	}()

	// read
	go func() {
		for {
			msg, err := packer.Unpack(conn)
			if err != nil {
				log.Error(err)
				return
			}
			log.Infof("sender | recv ack | %s", msg.Data())
		}
	}()
}

func readerClient(id int) {
	conn, err := establish()
	if err != nil {
		log.Error(err)
		return
	}

	go func() {
		for {
			msg, err := packer.Unpack(conn)
			if err != nil {
				log.Error(err)
				return
			}
			log.Debugf("reader %03d | recv broadcast | %s", id, msg.Data())
		}
	}()
}
