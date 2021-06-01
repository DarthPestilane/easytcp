package main

import (
	"fmt"
	"github.com/DarthPestilane/easytcp"
	"github.com/DarthPestilane/easytcp/examples/fixture"
	"github.com/DarthPestilane/easytcp/logger"
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/DarthPestilane/easytcp/router"
	"github.com/DarthPestilane/easytcp/server"
	"github.com/DarthPestilane/easytcp/session"
	"github.com/sirupsen/logrus"
	"runtime"
	"time"
)

var log *logrus.Logger

func init() {
	log = logger.Default
}

func main() {
	// go printGoroutineNum()

	s := easytcp.NewTcp(server.TcpOption{
		RWBufferSize: 1024 * 1024,
	})

	easytcp.RegisterRoute(fixture.MsgIdPingReq, handler, fixture.RecoverMiddleware(log), logMiddleware)

	go func() {
		if err := s.Serve(fixture.ServerAddr); err != nil {
			log.Errorf("serve err: %s", err)
		}
	}()

	if err := s.GracefulStop(); err != nil {
		panic(err)
	}

	time.Sleep(time.Second * 3)
}

func handler(s *session.Session, req *packet.Request) (*packet.Response, error) {
	var data string
	_ = s.MsgCodec.Decode(req.RawData, &data)

	panicMaker := map[bool]struct{}{
		true:  {},
		false: {},
	}
	for k := range panicMaker {
		if !k {
			panic("random panic here")
		}
		break
	}

	return &packet.Response{
		Id:   fixture.MsgIdPingAck,
		Data: data + "||pong, pong, pong",
	}, nil
}

func logMiddleware(next router.HandlerFunc) router.HandlerFunc {
	return func(s *session.Session, req *packet.Request) (*packet.Response, error) {
		var data string
		_ = s.MsgCodec.Decode(req.RawData, &data)
		log.Infof("recv req | id:(%d) size:(%d) data: %s", req.Id, req.RawSize, data)
		return next(s, req)
	}
}

// nolint: deadcode, unused
func printGoroutineNum() {
	for {
		fmt.Println("goroutine num: ", runtime.NumGoroutine())
		time.Sleep(time.Second)
	}
}
