package router

import (
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/DarthPestilane/easytcp/session"
	"sync"
	"time"
)

// RespKey is the preset key to the response data before encoding.
const RespKey = "easytcp.router.context.response"

// Context is a generic context in a message routing.
// It allows us to pass variables between handler and middlewares.
// Context implements the context.Context interface.
type Context struct {
	storage sync.Map
	session session.Session
	reqMsg  packet.Message
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

// MsgID returns the request message's ID.
func (c *Context) MsgID() uint {
	return c.reqMsg.GetID()
}

// MsgSize returns the request message's size.
func (c *Context) MsgSize() uint {
	return c.reqMsg.GetSize()
}

// MsgRawData returns the request message's data, which may been encoded.
func (c *Context) MsgRawData() []byte {
	return c.reqMsg.GetData()
}

// Bind binds the request message's raw data to v.
func (c *Context) Bind(v interface{}) error {
	return c.session.MsgCodec().Decode(c.MsgRawData(), v)
}

// SessionID returns current session's ID.
func (c *Context) SessionID() string {
	return c.session.ID()
}

// Response creates a response message.
func (c *Context) Response(id uint, data interface{}) (packet.Message, error) {
	c.Set(RespKey, data)
	dataRaw, err := c.session.MsgCodec().Encode(data)
	if err != nil {
		return nil, err
	}
	respMsg := c.reqMsg.Duplicate()
	respMsg.Setup(id, dataRaw)
	return respMsg, nil
}

func newContext(sess session.Session, msg packet.Message) *Context {
	return &Context{session: sess, reqMsg: msg}
}
