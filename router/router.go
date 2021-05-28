package router

import (
	"context"
	"fmt"
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
	mapper sync.Map // msgId : HandleFunc 的映射
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
func (r *Router) Loop(ctx context.Context, s *session.Session) error {
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context done: %s", ctx.Err())
		case req, ok := <-s.RecvReq():
			if !ok {
				r.log.Trace("loop stopped since session's closed")
				return fmt.Errorf("receive request err: channel closed")
			}
			if req != nil {
				go func() {
					if err := r.handleReq(s, req); err != nil {
						r.log.Errorf("handle request err: %s", err)
					}
				}()
			}
		}
	}
}

func (r *Router) handleReq(s *session.Session, req *packet.Request) error {
	if v, has := r.mapper.Load(req.Id); has {
		if handler, ok := v.(HandleFunc); ok {
			resp := handler(s, req)
			if resp == nil {
				return nil
			}
			if err := s.SendResp(resp); err != nil {
				return fmt.Errorf("session send response err: %s", err)
			}
			return nil
		}
	}
	return fmt.Errorf("handler not found")
}

// Register 注册路由
func (r *Router) Register(id uint, fn HandleFunc) {
	r.mapper.Store(id, fn)
}
