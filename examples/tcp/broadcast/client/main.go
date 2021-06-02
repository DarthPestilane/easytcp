package main

import (
	"fmt"
	"github.com/DarthPestilane/easytcp/examples/fixture"
	"github.com/DarthPestilane/easytcp/logger"
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

var log *logrus.Logger
var codec packet.Codec
var packer packet.Packer

func init() {
	log = logger.Default
	codec = &packet.StringCodec{}
	packer = &packet.DefaultPacker{}
}

func main() {
	senderClient()
	readerClient(1)
	readerClient(2)

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
			data, _ := codec.Encode(fmt.Sprintf("hello everyone @%d", time.Now().Unix()))
			msg, _ := packer.Pack(fixture.MsgIdBroadCastReq, data)
			if _, err := conn.Write(msg); err != nil {
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
			log.Infof("sender | recv ack | %s", msg.GetData())
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
			log.Debugf("reader %03d | recv broadcast | %s", id, msg.GetData())
		}
	}()
}
