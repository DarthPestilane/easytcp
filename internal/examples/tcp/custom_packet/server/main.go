package main

import (
	"fmt"
	"github.com/DarthPestilane/easytcp"
	"github.com/DarthPestilane/easytcp/internal/examples/fixture"
	"github.com/DarthPestilane/easytcp/internal/examples/tcp/custom_packet/common"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
	log.SetLevel(logrus.DebugLevel)
}

func main() {
	easytcp.SetLogger(log)

	s := easytcp.NewServer(&easytcp.ServerOption{
		// specify codec and packer
		Codec:  &easytcp.JsonCodec{},
		Packer: &common.CustomPacker{},
	})

	s.AddRoute("json01-req", handler, fixture.RecoverMiddleware(log), logMiddleware)

	go func() {
		if err := s.Run(fixture.ServerAddr); err != nil {
			log.Errorf("serve err: %s", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	if err := s.Stop(); err != nil {
		log.Errorf("server stopped err: %s", err)
	}
}

func handler(ctx easytcp.Context) {
	var data common.Json01Req
	_ = ctx.Bind(&data)

	err := ctx.SetResponse("json01-resp", &common.Json01Resp{
		Success: true,
		Data:    fmt.Sprintf("%s:%d:%t", data.Key1, data.Key2, data.Key3),
	})
	if err != nil {
		log.Errorf("set response failed: %s", err)
	}
}

func logMiddleware(next easytcp.HandlerFunc) easytcp.HandlerFunc {
	return func(ctx easytcp.Context) {
		fullSize := ctx.Request().MustGet("fullSize")
		req := ctx.Request()
		log.Infof("recv request  | fullSize:(%d) id:(%v) dataSize(%d) data: %s", fullSize, req.ID(), len(req.Data()), req.Data())

		defer func() {
			resp := ctx.Response()
			if resp != nil {
				log.Infof("send response | dataSize:(%d) id:(%v) data: %s", len(resp.Data()), resp.ID(), resp.Data())
			} else {
				log.Infof("don't send response since nil")
			}
		}()
		next(ctx)
	}
}
