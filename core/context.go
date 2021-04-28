package core

import (
	"github.com/DarthPestilane/easytcp/codec"
	"sync"
	"time"
)

type Context struct {
	mu        sync.Mutex
	store     map[string]interface{}
	conn      *Connection
	body      []byte
	length    int
	routePath string
}

func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return
}

func (c *Context) Done() <-chan struct{} {
	return nil
}

func (c *Context) Err() error {
	return nil
}

func (c *Context) Value(key interface{}) interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()
	if keyAsStr, ok := key.(string); ok {
		return c.store[keyAsStr]
	}
	return nil
}

func (c *Context) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[key] = value
}

func (c *Context) setConn(conn *Connection) *Context {
	c.conn = conn
	return c
}

func (c *Context) setBody(body []byte) *Context {
	c.body = body
	return c
}

func (c *Context) setLength(n int) *Context {
	c.length = n
	return c
}

func (c *Context) setRoutePath(path string) *Context {
	c.routePath = path
	return c
}

func (c *Context) Bind(codec codec.Codec, data interface{}) error {
	return codec.Unmarshal(c.Body(), data)
}

func (c *Context) Conn() *Connection {
	return c.conn
}

func (c *Context) Body() []byte {
	return c.body
}

func (c *Context) Length() int {
	return c.length
}

func (c *Context) RoutePath() string {
	return c.routePath
}

func NewContext() *Context {
	return &Context{
		store: make(map[string]interface{}),
	}
}
