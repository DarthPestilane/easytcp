package router

import (
	"fmt"
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
	reqMsg  *packet.MessageEntry
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
	return c.reqMsg.ID
}

// MsgSize returns the request message's size.
func (c *Context) MsgSize() int {
	return len(c.reqMsg.Data)
}

// MsgData returns the request message's data, which may been encoded.
func (c *Context) MsgData() []byte {
	return c.reqMsg.Data
}

// Bind binds the request message's raw data to v.
func (c *Context) Bind(v interface{}) error {
	codec := c.session.Codec()
	if codec == nil {
		return fmt.Errorf("message codec is nil")
	}
	return codec.Decode(c.MsgData(), v)
}

// SessionID returns current session's ID.
func (c *Context) SessionID() string {
	return c.session.ID()
}

// Response creates a response message.
func (c *Context) Response(id uint, data interface{}) (*packet.MessageEntry, error) {
	c.Set(RespKey, data)
	var dataRaw []byte
	if codec := c.session.Codec(); codec == nil {
		switch v := data.(type) {
		case []byte:
			dataRaw = v
		case string:
			dataRaw = []byte(v)
		case fmt.Stringer:
			dataRaw = []byte(v.String())
		default:
			return nil, fmt.Errorf("data should be []byte, string or Stringer")
		}
	} else {
		var err error
		dataRaw, err = codec.Encode(data)
		if err != nil {
			return nil, err
		}
	}
	respMsg := &packet.MessageEntry{
		ID:   id,
		Data: dataRaw,
	}
	return respMsg, nil
}

func newContext(sess session.Session, msg *packet.MessageEntry) *Context {
	return &Context{session: sess, reqMsg: msg}
}
