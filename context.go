package tcp_demo

import (
	"bytes"
	"demo/tcp_demo/codec"
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

func (c *Context) Write(content []byte) (int, error) {
	msg := bytes.TrimRight(content, "\n")
	return c.Conn().Write(append(msg, '\n'))
}

func (c *Context) WriteIn(content []byte, duration time.Duration) (int, error) {
	_ = c.Conn().SetWriteDeadline(time.Now().Add(duration))
	defer c.Conn().SetWriteDeadline(time.Time{}) // zero time
	return c.Write(content)
}

func (c *Context) WriteString(content string) (int, error) {
	return c.Write([]byte(content))
}

func (c *Context) WriteStringIn(content string, duration time.Duration) (int, error) {
	return c.WriteIn([]byte(content), duration)
}

func NewContext() *Context {
	return &Context{
		store: make(map[string]interface{}),
	}
}
