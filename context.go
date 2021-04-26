package tcp_demo

import (
	"demo/tcp_demo/codec"
	"demo/tcp_demo/util/message"
	"net"
	"sync"
	"time"
)

type Context struct {
	mu    sync.Mutex
	store map[string]interface{}
}

const (
	CtxKeyConn = "conn"
	CtxKeyBody = "body"
)

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

func (c *Context) setConn(conn net.Conn) *Context {
	c.Set(CtxKeyConn, conn)
	return c
}

func (c *Context) setBody(body []byte) *Context {
	c.Set(CtxKeyBody, body)
	return c
}

func (c *Context) Bind(codec codec.Codec, data interface{}) error {
	return codec.Unmarshal(c.Body(), data)
}

func (c *Context) Conn() net.Conn {
	return c.Value(CtxKeyConn).(net.Conn)
}

func (c *Context) Body() []byte {
	return c.Value(CtxKeyBody).([]byte)
}

func (c *Context) Send(routePath string, b []byte) (int, error) {
	msg := message.AddHead(routePath, b)
	msg = append(msg, '\n')
	return c.Conn().Write(msg)
}

func (c *Context) SendIn(routePath string, b []byte, duration time.Duration) (int, error) {
	msg := message.AddHead(routePath, b)
	msg = append(msg, '\n')

	if err := c.Conn().SetWriteDeadline(time.Now().Add(duration)); err != nil {
		return 0, err
	}
	defer c.Conn().SetWriteDeadline(time.Time{})
	return c.Conn().Write(msg)
}

func NewContext() *Context {
	return &Context{
		store: make(map[string]interface{}),
	}
}
