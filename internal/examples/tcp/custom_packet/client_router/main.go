package main

import (
	"github.com/DarthPestilane/easytcp"
	"github.com/DarthPestilane/easytcp/internal/examples/fixture"
	"github.com/DarthPestilane/easytcp/internal/examples/tcp/custom_packet/common"
	"github.com/sirupsen/logrus"
	"time"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
	log.SetLevel(logrus.DebugLevel)
}

func main() {
	notify := make(chan interface{})
	client := easytcp.NewClient(&easytcp.ClientOption{
		// specify codec and packer
		ServerOption: easytcp.ServerOption{
			Codec:  &easytcp.JsonCodec{},
			Packer: &common.CustomPacker{},
		},
		NotifyChan: notify,
	})
	err := client.Run(fixture.ServerAddr)
	if err != nil {
		panic(err)
	}
	client.AddRoute("json01-resp", respHandler, logMiddleware)
	log = logrus.New()
	go func() {
		// write loop
		i := 0
		for {
			time.Sleep(time.Second)
			req := &common.Json01Req{
				Key1: "hello",
				Key2: i,
				Key3: true,
			}
			err = client.Send("json01-req", req)
			if err != nil {
				panic(err)
			}
			i++
		}
	}()
	i := 0
	for true {
		if client.IsStopped() {
			log.Infof("stop")
			break
		}
		select {
		case v := <-notify:
			if i == 10 {
				_ = client.Stop()
			}
			i++
			log.Infof("recv notify %v", v)
		}
	}
}

func respHandler(ctx easytcp.Context) {
	var data common.Json01Resp
	_ = ctx.Bind(&data)
	ctx.Notify(data.Data)
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
			}
		}()
		next(ctx)
	}
}
