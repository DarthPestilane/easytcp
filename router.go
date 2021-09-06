package easytcp

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"os"
	"reflect"
	"runtime"
)

func newRouter(queueSize ...int) *Router {
	size := 0
	if len(queueSize) != 0 {
		if qs := queueSize[0]; qs > 0 {
			size = qs
		}
	}
	return &Router{
		reqQueue:          make(chan *Context, size),
		stopped:           make(chan struct{}),
		handlerMapper:     make(map[interface{}]HandlerFunc),
		middlewaresMapper: make(map[interface{}][]MiddlewareFunc),
	}
}

// Router is a router for incoming message.
// Router routes the message to its handler and middlewares.
type Router struct {
	// handlerMapper maps message's ID to handler.
	// Handler will be called around middlewares.
	handlerMapper map[interface{}]HandlerFunc

	// middlewaresMapper maps message's ID to a list of middlewares.
	// These middlewares will be called before the handler in handlerMapper.
	middlewaresMapper map[interface{}][]MiddlewareFunc

	// globalMiddlewares is a list of MiddlewareFunc.
	// globalMiddlewares will be called before the ones in middlewaresMapper.
	globalMiddlewares []MiddlewareFunc

	notFoundHandler HandlerFunc
	reqQueue        chan *Context
	stopped         chan struct{}
}

// HandlerFunc is the function type for handlers.
type HandlerFunc func(ctx *Context) error

// MiddlewareFunc is the function type for middlewares.
// A common pattern is like:
//
// 	var md MiddlewareFunc = func(next HandlerFunc) HandlerFunc {
// 		return func(ctx *Context) error {
// 			return next(ctx)
// 		}
// 	}
type MiddlewareFunc func(next HandlerFunc) HandlerFunc

var nilHandler HandlerFunc = func(ctx *Context) error {
	return nil
}

// stop stops the router.
func (r *Router) stop() {
	close(r.stopped)
}

// consumeRequest fetches context from reqQueue, and handle it.
func (r *Router) consumeRequest() {
	defer Log.Tracef("router stopped")
	for {
		select {
		case <-r.stopped:
			close(r.reqQueue)
			return
		case ctx, ok := <-r.reqQueue:
			if !ok {
				return
			}
			select {
			case <-ctx.session.closed:
				continue
			default:
			}

			go func() {
				if err := r.handleRequest(ctx); err != nil {
					Log.Errorf("router handle request err: %s", err)
					return
				}
				if err := ctx.session.SendResp(ctx); err != nil {
					Log.Errorf("router send resp err: %s", err)
				}
			}()
		}
	}
}

// handleRequest walks ctx through middlewares and handler,
// and returns response message entry.
func (r *Router) handleRequest(ctx *Context) error {
	if ctx.reqEntry == nil {
		return nil
	}
	var handler HandlerFunc
	if v, has := r.handlerMapper[ctx.reqEntry.ID]; has {
		handler = v
	}

	var mws = r.globalMiddlewares
	if v, has := r.middlewaresMapper[ctx.reqEntry.ID]; has {
		mws = append(mws, v...) // append to global ones
	}

	// create the handlers stack
	wrapped := r.wrapHandlers(handler, mws)

	// and call the handlers stack
	return wrapped(ctx)
}

// wrapHandlers wraps handler and middlewares into a right order call stack.
// Makes something like:
// 	var wrapped HandlerFunc = m1(m2(m3(handle)))
func (r *Router) wrapHandlers(handler HandlerFunc, middles []MiddlewareFunc) (wrapped HandlerFunc) {
	if handler == nil {
		if r.notFoundHandler != nil {
			handler = r.notFoundHandler
		} else {
			handler = nilHandler
		}
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
		r.handlerMapper[id] = h
	}
	ms := make([]MiddlewareFunc, 0, len(m))
	for _, mm := range m {
		if mm != nil {
			ms = append(ms, mm)
		}
	}
	if len(ms) != 0 {
		r.middlewaresMapper[id] = ms
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
	for id, h := range r.handlerMapper {
		handlerName := runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
		table.Append([]string{fmt.Sprintf("%v", id), handlerName})
	}
	table.Render()
	fmt.Printf("[EASYTCP] Serving at: %s\n\n", addr)
}

func (r *Router) setNotFoundHandler(handler HandlerFunc) {
	r.notFoundHandler = handler
}
