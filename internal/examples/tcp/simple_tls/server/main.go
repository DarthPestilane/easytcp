package main

import (
	"crypto/tls"
	"fmt"
	"github.com/DarthPestilane/easytcp"
	"github.com/DarthPestilane/easytcp/internal/examples/fixture"
	"github.com/DarthPestilane/easytcp/internal/examples/tcp/simple_tls/common"
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
		ReadTimeout:   time.Second * 3,
		WriteTimeout:  time.Second * 3,
		RespQueueSize: -1,
		Packer:        easytcp.NewDefaultPacker(),
		Codec:         nil,
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
		c.SetResponseMessage(easytcp.NewMessage(common.MsgIdPingAck, []byte("pong, pong, pong")))
	})

	cert, err := tls.LoadX509KeyPair("internal/test_data/certificates/cert.pem", "internal/test_data/certificates/cert.key")
	if err != nil {
		panic(err)
	}
	go func() {
		if err := s.RunTLS(fixture.ServerAddr, &tls.Config{Certificates: []tls.Certificate{cert}}); err != nil {
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
		log.Infof("rec <<< id:(%d) size:(%d) data: %s", req.ID, len(req.Data()), req.Data())
		defer func() {
			resp := c.Response()
			log.Infof("snd >>> id:(%d) size:(%d) data: %s", resp.ID, len(resp.Data()), resp.Data())
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
