package main

import (
	"fmt"
	"github.com/DarthPestilane/easytcp"
	"github.com/DarthPestilane/easytcp/internal/examples/fixture"
	"github.com/DarthPestilane/easytcp/internal/examples/tcp/broadcast/common"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var log *logrus.Logger
var sessions *SessionManager

func init() {
	log = logrus.New()
	sessions = &SessionManager{nextId: 1, storage: map[int64]easytcp.Session{}}
}

type SessionManager struct {
	nextId  int64
	lock    sync.Mutex
	storage map[int64]easytcp.Session
}

func main() {
	s := easytcp.NewServer(&easytcp.ServerOption{
		Packer: easytcp.NewDefaultPacker(),
	})

	s.OnSessionCreate = func(sess easytcp.Session) {
		// store session
		sessions.lock.Lock()
		defer sessions.lock.Unlock()
		sess.SetID(sessions.nextId)
		sessions.nextId++
		sessions.storage[sess.ID().(int64)] = sess
	}

	s.OnSessionClose = func(sess easytcp.Session) {
		// remove session
		delete(sessions.storage, sess.ID().(int64))
	}

	s.Use(fixture.RecoverMiddleware(log), logMiddleware)

	s.AddRoute(common.MsgIdBroadCastReq, func(ctx easytcp.Context) {
		reqData := ctx.Request().Data()

		// broadcasting to other sessions
		currentSession := ctx.Session()
		for _, sess := range sessions.storage {
			targetSession := sess
			if currentSession.ID() == targetSession.ID() {
				continue
			}
			respData := fmt.Sprintf("%s (broadcast from %d to %d)", reqData, currentSession.ID(), targetSession.ID())
			respMsg := easytcp.NewMessage(common.MsgIdBroadCastAck, []byte(respData))
			go func() {
				targetSession.AllocateContext().SetResponseMessage(respMsg).Send()
				// can also write like this.
				// ctx.Copy().SetResponseMessage(respMsg).SendTo(targetSession)
				// or this.
				// ctx.Copy().SetSession(targetSession).SetResponseMessage(respMsg).Send()
			}()
		}

		ctx.SetResponseMessage(easytcp.NewMessage(common.MsgIdBroadCastAck, []byte("broadcast done")))
	})

	go func() {
		if err := s.Run(fixture.ServerAddr); err != nil {
			log.Error(err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	if err := s.Stop(); err != nil {
		log.Errorf("server stopped err: %s", err)
	}
	time.Sleep(time.Second)
}

func logMiddleware(next easytcp.HandlerFunc) easytcp.HandlerFunc {
	return func(ctx easytcp.Context) {
		log.Infof("recv request | %s", ctx.Request().Data())
		defer func() {
			var respMsg = ctx.Response()
			log.Infof("send response |sessId: %d; id: %d; size: %d; data: %s", ctx.Session().ID(), respMsg.ID(), len(respMsg.Data()), respMsg.Data())
		}()
		next(ctx)
	}
}
