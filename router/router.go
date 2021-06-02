package router

import (
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
	log               *logrus.Entry
	handlerMapper     sync.Map
	middlewaresMapper sync.Map
	globalMiddlewares []MiddlewareFunc
}

type HandlerFunc func(s session.Session, req *packet.Request) (*packet.Response, error)
type MiddlewareFunc func(next HandlerFunc) HandlerFunc

var defaultHandler HandlerFunc = func(s session.Session, req *packet.Request) (*packet.Response, error) {
	return nil, nil
}

// Instance 单例模式，获取 *Router 实例(instance)
func Instance() *Router {
	once.Do(func() {
		rt = &Router{
			log: logger.Default.WithField("scope", "router.Router"),
		}
	})
	return rt
}

// Loop 阻塞式消费 session.Session 中的 reqQueue channel
// 通过消息ID找到对应的 HandleFunc 并调用
func (r *Router) Loop(s session.Session) error {
	for {
		req, ok := <-s.RecvReq()
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

func (r *Router) handleReq(s session.Session, req *packet.Request) error {
	var handler HandlerFunc
	if v, has := r.handlerMapper.Load(req.Id); has {
		handler = v.(HandlerFunc)
	}

	var middles = r.globalMiddlewares
	if v, has := r.middlewaresMapper.Load(req.Id); has {
		middles = append(middles, v.([]MiddlewareFunc)...) // append to global ones
	}

	wrapped := r.wrapHandlers(handler, middles)

	// call the handlers stack now
	resp, err := wrapped(s, req)
	if err != nil {
		return fmt.Errorf("handler err: %s", err)
	}
	if resp == nil {
		return nil
	}
	if err := s.SendResp(resp); err != nil {
		return fmt.Errorf("session send response err: %s", err)
	}
	return nil
}

// wrapHandlers make something like wrapped = M1(M2(M3(handle)))
func (r *Router) wrapHandlers(handler HandlerFunc, middles []MiddlewareFunc) (wrapped HandlerFunc) {
	if handler == nil {
		handler = defaultHandler
	}
	wrapped = handler
	for i := len(middles) - 1; i >= 0; i-- {
		m := middles[i]
		wrapped = m(wrapped)
	}
	return wrapped
}

// Register 注册路由
func (r *Router) Register(id uint, h HandlerFunc, m ...MiddlewareFunc) {
	if h != nil {
		r.handlerMapper.Store(id, h)
	}
	if len(m) != 0 {
		r.middlewaresMapper.Store(id, m)
	}
}

// RegisterMiddleware 注册全局中间件
func (r *Router) RegisterMiddleware(m ...MiddlewareFunc) {
	if len(m) != 0 {
		r.globalMiddlewares = append(r.globalMiddlewares, m...)
	}
}
