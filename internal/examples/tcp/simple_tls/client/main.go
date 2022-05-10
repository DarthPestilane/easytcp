package main

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/DarthPestilane/easytcp"
	"github.com/DarthPestilane/easytcp/internal/examples/fixture"
	"github.com/DarthPestilane/easytcp/internal/examples/tcp/simple_tls/common"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"time"
)

func main() {
	pem, err := ioutil.ReadFile("internal/test_data/certificates/cert.pem")
	if err != nil {
		panic(err)
	}
	rootCerts := x509.NewCertPool()
	rootCerts.AppendCertsFromPEM(pem)
	tlsCfg := &tls.Config{
		RootCAs:            rootCerts,
		InsecureSkipVerify: true,
	}
	conn, err := tls.Dial("tcp", fixture.ServerAddr, tlsCfg)
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
