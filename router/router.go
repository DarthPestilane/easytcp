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

// Router 路由器，负责对消息的路由
type Router struct {
	// mapper msgId -> HandleFunc 的映射
	mapper sync.Map
	log    *logrus.Entry
}

type HandleFunc func(s *session.Session, req *packet.Request) *packet.Response

// Inst 单例模式，获取 *Router 实例(instance)
func Inst() *Router {
	once.Do(func() {
		rt = &Router{
			log: logger.Default.WithField("scope", "router.Router"),
		}
	})
	return rt
}

// Loop 阻塞式消费 session.Session 中的 reqQueue channel
// 通过消息ID找到对应的 HandleFunc 并调用
func (r *Router) Loop(s *session.Session) {
	for {
		req, ok := s.RecvReq()
		if !ok {
			r.log.Warnf("session closed. loop finished")
			return
		}
		if req != nil {
			go r.handleReq(s, req)
		}
	}
}

func (r *Router) handleReq(s *session.Session, req *packet.Request) {
	if v, has := r.mapper.Load(req.Id); has {
		if handler, ok := v.(HandleFunc); ok {
			resp := handler(s, req)
			if err := s.SendResp(resp); err != nil {
				r.log.Errorf("session send resp err: %s", err)
			}
		}
	}
}

// Register 注册路由
func (r *Router) Register(id uint, fn HandleFunc) {
	r.mapper.Store(id, fn)
}
