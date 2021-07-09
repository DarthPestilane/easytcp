package easytcp

import (
	"fmt"
	"github.com/DarthPestilane/easytcp/message"
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
	session *Session
	reqMsg  *message.Entry
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

// Message returns the request message entry.
func (c *Context) Message() *message.Entry {
	return c.reqMsg
}

// Bind binds the request message's raw data to v.
func (c *Context) Bind(v interface{}) error {
	codec := c.session.codec
	if codec == nil {
		return fmt.Errorf("message codec is nil")
	}
	return codec.Decode(c.reqMsg.Data, v)
}

// Session returns current session.
func (c *Context) Session() *Session {
	return c.session
}

// Response creates a response message.
func (c *Context) Response(id interface{}, data interface{}) (*message.Entry, error) {
	c.Set(RespKey, data)
	var dataRaw []byte
	if codec := c.session.codec; codec == nil {
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
	respMsg := &message.Entry{
		ID:   id,
		Data: dataRaw,
	}
	return respMsg, nil
}
