package easytcp

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cast"
	"io"
	"os"
	"reflect"
	"runtime"
	"sort"
)

func newRouter() *Router {
	return &Router{
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
}

// HandlerFunc is the function type for handlers.
type HandlerFunc func(ctx Context)

// MiddlewareFunc is the function type for middlewares.
// A common pattern is like:
//
// 	var mf MiddlewareFunc = func(next HandlerFunc) HandlerFunc {
// 		return func(ctx Context) {
// 			next(ctx)
// 		}
// 	}
type MiddlewareFunc func(next HandlerFunc) HandlerFunc

var nilHandler HandlerFunc = func(ctx Context) {}

// handleRequest walks ctx through middlewares and handler,
// and returns response message entry.
func (r *Router) handleRequest(ctx Context) {
	reqEntry := ctx.Request()
	if reqEntry == nil {
		return
	}
	var handler HandlerFunc
	if v, has := r.handlerMapper[reqEntry.ID]; has {
		handler = v
	}

	var mws = r.globalMiddlewares
	if v, has := r.middlewaresMapper[reqEntry.ID]; has {
		mws = append(mws, v...) // append to global ones
	}

	// create the handlers stack
	wrapped := r.wrapHandlers(handler, mws)

	// and call the handlers stack
	wrapped(ctx)
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
	var w io.Writer = os.Stdout

	_, _ = fmt.Fprintf(w, "\n[EASYTCP] Message-Route Table: \n")
	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Message ID", "Route Handler"})
	table.SetAutoFormatHeaders(false)

	// sort ids
	ids := make([]interface{}, 0, len(r.handlerMapper))
	for id := range r.handlerMapper {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		a, b := cast.ToString(ids[i]), cast.ToString(ids[j])
		return a < b
	})

	// add table row
	for _, id := range ids {
		h := r.handlerMapper[id]
		handlerName := runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
		table.Append([]string{fmt.Sprintf("%v", id), handlerName})
	}

	table.Render()
	_, _ = fmt.Fprintf(w, "[EASYTCP] Serving at: %s\n\n", addr)
}

func (r *Router) setNotFoundHandler(handler HandlerFunc) {
	r.notFoundHandler = handler
}
