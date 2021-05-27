package router

import (
	"github.com/DarthPestilane/easytcp/logger"
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/DarthPestilane/easytcp/session"
	"github.com/sirupsen/logrus"
	"sync"
)

var (
	once sync.Once
	rt   *Router
)

type Router struct {
	mapper sync.Map
	log    *logrus.Entry
}

type HandleFunc func(s *session.Session, msg *packet.Request)

func Inst() *Router {
	once.Do(func() {
		rt = &Router{
			log: logger.Default.WithField("scope", "router"),
		}
	})
	return rt
}

func (r *Router) Loop(s *session.Session) {
	for {
		req := s.RecvReq()
		if req == nil {
			// session closed
			r.log.Warnf("loop finished")
			return
		}
		r.log.Debugf("req: %+v", req)

		if v, has := r.mapper.Load(uint32(1)); has {
			if handler, ok := v.(HandleFunc); ok {
				r.log.Debugf("found handler")
				go handler(s, req)
			}
		}
	}
}

func (r *Router) Register(id uint32, fn HandleFunc) {
	r.mapper.Store(id, fn)
}
