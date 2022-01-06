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
	log.SetLevel(logrus.TraceLevel)
}

func main() {
	// go printGoroutineNum()

	easytcp.SetLogger(log)
	s := easytcp.NewServer(&easytcp.ServerOption{
		SocketReadBufferSize:  1024 * 1024,
		SocketWriteBufferSize: 1024 * 1024,
		ReadTimeout:           time.Second * 3,
		WriteTimeout:          time.Second * 3,
		RespQueueSize:         -1,
		Packer:                easytcp.NewDefaultPacker(),
		Codec:                 nil,
	})
	s.OnSessionCreate = func(sess easytcp.Session) {
		log.Infof("session created: %v", sess.ID())
	}
	s.OnSessionClose = func(sess easytcp.Session) {
		log.Warnf("session closed: %v", sess.ID())
	}

	// register global middlewares
	s.Use(fixture.RecoverMiddleware(log), logMiddleware)

	// register a route
	s.AddRoute(common.MsgIdPingReq, func(c easytcp.Context) {
		c.SetResponseMessage(&message.Entry{
			ID:   common.MsgIdPingAck,
			Data: []byte("pong, pong, pong"),
		})
	})

	go func() {
		if err := s.Serve(fixture.ServerAddr); err != nil {
			log.Errorf("serve err: %s", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	if err := s.Stop(); err != nil {
		log.Errorf("server stopped err: %s", err)
	}
	time.Sleep(time.Second * 3)
}

func logMiddleware(next easytcp.HandlerFunc) easytcp.HandlerFunc {
	return func(c easytcp.Context) {
		req := c.Request()
		log.Infof("rec <<< id:(%d) size:(%d) data: %s", req.ID, len(req.Data), req.Data)
		defer func() {
			resp := c.Response()
			log.Infof("snd >>> id:(%d) size:(%d) data: %s", resp.ID, len(resp.Data), resp.Data)
		}()
		next(c)
	}
}

// nolint: deadcode, unused
func printGoroutineNum() {
	for {
		fmt.Println("goroutine num: ", runtime.NumGoroutine())
		time.Sleep(time.Second)
	}
}
