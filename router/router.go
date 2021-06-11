package router

import (
	"fmt"
	"github.com/DarthPestilane/easytcp/logger"
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/DarthPestilane/easytcp/session"
	"github.com/sirupsen/logrus"
	"sync"
)

// Router is a router for incoming message.
// Router routes the message to its handler and middlewares.
type Router struct {
	log *logrus.Entry

	// handlerMapper maps message's ID to handler.
	// Handler will be called around middlewares.
	handlerMapper sync.Map

	// middlewaresMapper maps message's ID to a list of middlewares.
	// These middlewares will be called before the handler in handlerMapper.
	middlewaresMapper sync.Map

	// globalMiddlewares is a list of MiddlewareFunc.
	// globalMiddlewares will be called before the ones in middlewaresMapper.
	globalMiddlewares []MiddlewareFunc
}

// HandlerFunc is the function type for handlers.
type HandlerFunc func(ctx *Context) (packet.Message, error)

// MiddlewareFunc is the function type for middlewares.
// A common pattern is like:
//
// 	var md MiddlewareFunc = func(next HandlerFunc) HandlerFunc {
// 		return func(ctx *Context) (*packet.Response, error) {
// 			return next(ctx)
// 		}
// 	}
type MiddlewareFunc func(next HandlerFunc) HandlerFunc

var defaultHandler HandlerFunc = func(ctx *Context) (packet.Message, error) {
	return nil, nil
}

// NewRouter creates a new Router pointer.
func NewRouter() *Router {
	return &Router{
		log:               logger.Default.WithField("scope", "router.Router"),
		globalMiddlewares: make([]MiddlewareFunc, 0),
	}
}

// RouteLoop reads message from session.Session s in a loop way,
// and routes the message to corresponding handler and middlewares if message is not nil.
// RouteLoop will break if session.Session s is closed.
func (r *Router) RouteLoop(s session.Session) {
	for {
		req, ok := <-s.RecvReq()
		if !ok {
			r.log.WithField("sid", s.ID()).Tracef("loop stopped since session is closed")
			break
		}
		if req == nil {
			continue
		}
		go func() {
			if err := r.handleReq(s, req); err != nil {
				r.log.WithField("sid", s.ID()).Tracef("handle request err: %s", err)
			}
		}()
	}
	r.log.WithField("sid", s.ID()).Tracef("loop exit")
}

// handleReq routes the packet.Message reqMsg to corresponding handler and middlewares,
// and call the handler functions, and send response to session.Session s if response is not nil.
// Returns error when calling handler functions or sending response failed.
func (r *Router) handleReq(s session.Session, reqMsg packet.Message) error {
	var handler HandlerFunc
	if v, has := r.handlerMapper.Load(reqMsg.GetID()); has {
		handler = v.(HandlerFunc)
	}

	var mws = r.globalMiddlewares
	if v, has := r.middlewaresMapper.Load(reqMsg.GetID()); has {
		mws = append(mws, v.([]MiddlewareFunc)...) // append to global ones
	}

	// create context
	ctx := newContext(s, reqMsg)

	// create the handlers stack
	wrapped := r.wrapHandlers(handler, mws)

	// call the handlers stack now
	resp, err := wrapped(ctx)
	if err != nil {
		return fmt.Errorf("handler err: %s", err)
	}
	if resp == nil {
		return nil
	}
	if _, err := s.SendResp(resp); err != nil {
		return fmt.Errorf("session send response err: %s", err)
	}
	return nil
}

// wrapHandlers make something like wrapped = M1(M2(M3(handle))).
// wrapHandlers wraps handler and middlewares into a right order call stack.
// Makes something like:
// 	var wrapped HandlerFunc = m1(m2(m3(handle)))
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

// Register stores handler and middlewares for id.
func (r *Router) Register(id uint, h HandlerFunc, m ...MiddlewareFunc) {
	if h != nil {
		r.handlerMapper.Store(id, h)
	}
	if len(m) != 0 {
		ms := make([]MiddlewareFunc, 0)
		for _, mm := range m {
			if mm != nil {
				ms = append(ms, mm)
			}
		}
		if len(ms) != 0 {
			r.middlewaresMapper.Store(id, ms)
		}
	}
}

// RegisterMiddleware stores the global middlewares.
func (r *Router) RegisterMiddleware(m ...MiddlewareFunc) {
	if len(m) != 0 {
		for _, mm := range m {
			if mm != nil {
				r.globalMiddlewares = append(r.globalMiddlewares, mm)
			}
		}
	}
}
