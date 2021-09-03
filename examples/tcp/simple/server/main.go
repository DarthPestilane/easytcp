package main

import (
	"fmt"
	"github.com/DarthPestilane/easytcp"
	"github.com/DarthPestilane/easytcp/examples/fixture"
	"github.com/DarthPestilane/easytcp/examples/tcp/simple/common"
	"github.com/DarthPestilane/easytcp/message"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
}

func main() {
	// go printGoroutineNum()

	s := easytcp.NewServer(&easytcp.ServerOption{
		SocketReadBufferSize:  1024 * 1024,
		SocketWriteBufferSize: 1024 * 1024,
		ReadTimeout:           time.Second * 3,
		WriteTimeout:          time.Second * 3,
		ReqQueueSize:          -1,
		RespQueueSize:         -1,
		Packer:                easytcp.NewDefaultPacker(),
		Codec:                 nil,
	})
	s.OnSessionCreate = func(sess *easytcp.Session) {
		log.Infof("session created: %s", sess.ID())
	}
	s.OnSessionClose = func(sess *easytcp.Session) {
		log.Warnf("session closed: %s", sess.ID())
	}

	// register global middlewares
	s.Use(fixture.RecoverMiddleware(log), logMiddleware)

	// register a route
	s.AddRoute(common.MsgIdPingReq, func(c *easytcp.Context) (*message.Entry, error) {
		return c.Response(common.MsgIdPingAck, "pong, pong, pong")
	})

	go func() {
		if err := s.Serve(fixture.ServerAddr); err != nil {
			log.Errorf("serve err: %s", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	<-sigCh
	if err := s.Stop(); err != nil {
		log.Errorf("server stopped err: %s", err)
	}
	time.Sleep(time.Second)
}

func logMiddleware(next easytcp.HandlerFunc) easytcp.HandlerFunc {
	return func(c *easytcp.Context) (resp *message.Entry, err error) {
		log.Infof("rec <<< | id:(%d) size:(%d) data: %s", c.Message().ID, len(c.Message().Data), c.Message().Data)
		defer func() {
			if err != nil || resp == nil {
				return
			}
			log.Infof("snd >>> | id:(%d) size:(%d) data: %s", resp.ID, len(resp.Data), resp.Data)
		}()
		return next(c)
	}
}

// nolint: deadcode, unused
func printGoroutineNum() {
	for {
		fmt.Println("goroutine num: ", runtime.NumGoroutine())
		time.Sleep(time.Second)
	}
}
