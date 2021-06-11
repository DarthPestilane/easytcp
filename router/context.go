package router

import (
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/DarthPestilane/easytcp/session"
	"sync"
	"time"
)

// Context is a generic context in a message routing.
// It allows us to pass variables between handler and middlewares.
// Context implements the context.Context interface.
type Context struct {
	storage sync.Map

	Session session.Session
	Message packet.Message
}

// Deadline implements the context.Context Deadline method.
func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return
}

// Done implements the context.Context Done method.
func (c *Context) Done() <-chan struct{} {
	return nil
}

// Err implements the context.Context Err method.
func (c *Context) Err() error {
	return nil
}

// Value implements the context.Context Value method.
// Value returns the value of c.storage by key.
func (c *Context) Value(key interface{}) interface{} {
	if keyAsString, ok := key.(string); ok {
		val, _ := c.Get(keyAsString)
		return val
	}
	return nil
}

// Get returns the value from c.storage by key.
func (c *Context) Get(key string) (interface{}, bool) {
	return c.storage.Load(key)
}

// Set sets the value in c.storage.
func (c *Context) Set(key string, value interface{}) {
	c.storage.Store(key, value)
}

func (c *Context) MessageID() uint {
	return c.Message.GetID()
}

func (c *Context) MessageSize() uint {
	return c.Message.GetSize()
}

func (c *Context) MessageRawData() []byte {
	return c.Message.GetData()
}

// Bind binds the message data to v.
// Returns error if occurred.
func (c *Context) Bind(v interface{}) error {
	return c.Session.MsgCodec().Decode(c.MessageRawData(), v)
}

func newContext(sess session.Session, msg packet.Message) *Context {
	return &Context{Session: sess, Message: msg}
}
