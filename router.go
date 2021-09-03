package easytcp

import (
	"fmt"
	"github.com/DarthPestilane/easytcp/message"
	"github.com/olekukonko/tablewriter"
	"os"
	"reflect"
	"runtime"
	"sync"
)

func newRouter(queueSize ...int) *Router {
	size := 0
	if len(queueSize) != 0 {
		if qs := queueSize[0]; qs > 0 {
			size = qs
		}
	}
	return &Router{
		reqCtxQueue: make(chan *Context, size),
		stopped:     make(chan struct{}),
	}
}

// Router is a router for incoming message.
// Router routes the message to its handler and middlewares.
type Router struct {
	// handlerMapper maps message's ID to handler.
	// Handler will be called around middlewares.
	handlerMapper sync.Map

	// middlewaresMapper maps message's ID to a list of middlewares.
	// These middlewares will be called before the handler in handlerMapper.
	middlewaresMapper sync.Map

	// globalMiddlewares is a list of MiddlewareFunc.
	// globalMiddlewares will be called before the ones in middlewaresMapper.
	globalMiddlewares []MiddlewareFunc

	notFoundHandler HandlerFunc
	reqCtxQueue     chan *Context
	stopped         chan struct{}
}

// HandlerFunc is the function type for handlers.
type HandlerFunc func(ctx *Context) (*message.Entry, error)

// MiddlewareFunc is the function type for middlewares.
// A common pattern is like:
//
// 	var md MiddlewareFunc = func(next HandlerFunc) HandlerFunc {
// 		return func(ctx *Context) (message.Entry, error) {
// 			return next(ctx)
// 		}
// 	}
type MiddlewareFunc func(next HandlerFunc) HandlerFunc

var nilHandler HandlerFunc = func(ctx *Context) (*message.Entry, error) {
	return nil, nil
}

func (r *Router) stop() {
	close(r.stopped)
}

func (r *Router) consumeRequest() {
	for {
		select {
		case <-r.stopped:
			close(r.reqCtxQueue)

			Log.Tracef("router stopped")
			return
		case reqCtx, ok := <-r.reqCtxQueue:
			if !ok {
				Log.Tracef("router stopped")
				return
			}
			select {
			case <-reqCtx.session.closed:
				continue
			default:
			}

			go func() {
				respEntry, err := r.handleRequest(reqCtx)
				if err != nil {
					Log.Errorf("router handle request err: %s", err)
					return
				}
				if respEntry == nil {
					return
				}
				if err := reqCtx.session.SendResp(respEntry); err != nil {
					Log.Errorf("router send resp err: %s", err)
				}
			}()
		}
	}
}

func (r *Router) handleRequest(ctx *Context) (*message.Entry, error) {
	var handler HandlerFunc
	if v, has := r.handlerMapper.Load(ctx.reqMsg.ID); has {
		handler = v.(HandlerFunc)
	}

	var mws = r.globalMiddlewares
	if v, has := r.middlewaresMapper.Load(ctx.reqMsg.ID); has {
		mws = append(mws, v.([]MiddlewareFunc)...) // append to global ones
	}

	// create the handlers stack
	wrapped := r.wrapHandlers(handler, mws)

	// and call the handlers stack
	return wrapped(ctx)
}

// routeLoop reads message from session.Session s in a loop way,
// and routes the message to corresponding handler and middlewares if message is not nil.
// routeLoop will break if session.Session s is closed.
func (r *Router) routeLoop(s *Session) {
	for {
		select {
		case <-s.closed:
			Log.Tracef("router loop exit because session is closed")
			return
		case req, ok := <-s.reqQueue:
			if !ok {
				Log.Tracef("router loop exit because session is closed")
				return
			}
			if req == nil {
				continue
			}
			go func() {
				resp, err := r.handleReq(s, req)
				if err != nil {
					Log.Errorf("router handle request err: %s", err)
					return
				}
				if resp == nil {
					return
				}
				if err := s.SendResp(resp); err != nil {
					Log.Errorf("router send resp err: %s", err)
				}
			}()
		}
	}
}

// handleReq routes the packet.Message reqMsg to corresponding handler and middlewares,
// and call the handler functions, and send response to session.Session s if response is not nil.
// Returns error when calling handler functions or sending response failed.
func (r *Router) handleReq(s *Session, reqMsg *message.Entry) (*message.Entry, error) {
	var handler HandlerFunc
	if v, has := r.handlerMapper.Load(reqMsg.ID); has {
		handler = v.(HandlerFunc)
	}

	var mws = r.globalMiddlewares
	if v, has := r.middlewaresMapper.Load(reqMsg.ID); has {
		mws = append(mws, v.([]MiddlewareFunc)...) // append to global ones
	}

	// create context
	ctx := &Context{session: s, reqMsg: reqMsg}

	// create the handlers stack
	wrapped := r.wrapHandlers(handler, mws)

	// call the handlers stack now
	return wrapped(ctx)
	// resp, err := wrapped(ctx)
	// if err != nil {
	// 	return fmt.Errorf("handler err: %s", err)
	// }
	// if resp == nil {
	// 	return nil
	// }
	// return s.SendResp(resp)
}

// wrapHandlers wraps handler and middlewares into a right order call stack.
// Makes something like:
// 	var wrapped HandlerFunc = m1(m2(m3(handle)))
func (r *Router) wrapHandlers(handler HandlerFunc, middles []MiddlewareFunc) (wrapped HandlerFunc) {
	if handler == nil {
		handler = r.notFoundHandler
	}
	if handler == nil {
		handler = nilHandler
	}
	wrapped = handler
	for i := len(middles) - 1; i >= 0; i-- {
		m := middles[i]
		wrapped = m(wrapped)
	}
	return wrapped
}

// register stores handler and middlewares for id.
func (r *Router) register(id interface{}, h HandlerFunc, m ...MiddlewareFunc) {
	if h != nil {
		r.handlerMapper.Store(id, h)
	}
	ms := make([]MiddlewareFunc, 0, len(m))
	for _, mm := range m {
		if mm != nil {
			ms = append(ms, mm)
		}
	}
	if len(ms) != 0 {
		r.middlewaresMapper.Store(id, ms)
	}
}

// registerMiddleware stores the global middlewares.
func (r *Router) registerMiddleware(m ...MiddlewareFunc) {
	for _, mm := range m {
		if mm != nil {
			r.globalMiddlewares = append(r.globalMiddlewares, mm)
		}
	}
}

// printHandlers prints registered route handlers to console.
func (r *Router) printHandlers(addr string) {
	fmt.Printf("\n[EASYTCP ROUTE TABLE]:\n")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Message ID", "Route Handler"})
	table.SetAutoFormatHeaders(false)
	r.handlerMapper.Range(func(key, value interface{}) bool {
		id := key
		handlerName := runtime.FuncForPC(reflect.ValueOf(value.(HandlerFunc)).Pointer()).Name()
		table.Append([]string{fmt.Sprintf("%v", id), handlerName})
		return true
	})
	table.Render()
	fmt.Printf("[EASYTCP] Serving at: %s\n\n", addr)
}

func (r *Router) setNotFoundHandler(handler HandlerFunc) {
	r.notFoundHandler = handler
}
