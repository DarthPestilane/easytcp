package easytcp

import (
	"context"
	"fmt"
	"github.com/DarthPestilane/easytcp/message"
	"sync"
	"time"
)

// NewContext creates a routeContext pointer.
func NewContext() *routeContext {
	return &routeContext{}
}

// Context is a generic context in a message routing.
// It allows us to pass variables between handler and middlewares.
type Context interface {
	context.Context

	// Session returns the current session.
	Session() Session

	// SetSession sets session.
	SetSession(sess Session)

	// Request returns request message entry.
	Request() *message.Entry

	// SetRequest sets request by id and data.
	SetRequest(id, data interface{}) error

	// SetRequestMessage sets request message entry directly.
	SetRequestMessage(entry *message.Entry) Context

	// Bind decodes request message entry to v.
	Bind(v interface{}) error

	// Response returns the response message entry.
	Response() *message.Entry

	// RawResponseData returns the not yet encoded response data.
	RawResponseData() interface{}

	// SetResponse encodes data with session's codec and sets response message entry.
	SetResponse(id, data interface{}) error

	// MustSetResponse encodes data with session's codec and sets response message entry.
	// panics on error.
	MustSetResponse(id, data interface{}) Context

	// SetResponseMessage sets response message entry directly.
	SetResponseMessage(entry *message.Entry) Context

	// Send sends itself to current session.
	Send()

	// SendTo sends itself to session.
	SendTo(session Session)

	// Get returns key value from storage.
	Get(key string) (value interface{}, exists bool)

	// Set store key value into storage.
	Set(key string, value interface{})

	// Remove deletes the key from storage.
	Remove(key string)

	// Copy returns a copy of Context.
	Copy() Context
}

// routeContext implements the Context interface.
type routeContext struct {
	mu          sync.RWMutex
	storage     map[string]interface{}
	session     Session
	reqEntry    *message.Entry
	respEntry   *message.Entry
	rawRespData interface{}
}

// Deadline implements the context.Context Deadline method.
func (c *routeContext) Deadline() (deadline time.Time, ok bool) {
	return
}

// Done implements the context.Context Done method.
func (c *routeContext) Done() <-chan struct{} {
	return nil
}

// Err implements the context.Context Err method.
func (c *routeContext) Err() error {
	return nil
}

// Value implements the context.Context Value method.
func (c *routeContext) Value(key interface{}) interface{} {
	if keyAsString, ok := key.(string); ok {
		val, _ := c.Get(keyAsString)
		return val
	}
	return nil
}

// Session implements Context.Session method.
func (c *routeContext) Session() Session {
	return c.session
}

// SetSession sets session.
func (c *routeContext) SetSession(sess Session) {
	c.session = sess
}

// Request implements Context.Request method.
func (c *routeContext) Request() *message.Entry {
	return c.reqEntry
}

// SetRequest sets request by id and data.
func (c *routeContext) SetRequest(id, data interface{}) error {
	codec := c.session.Codec()
	if codec == nil {
		return fmt.Errorf("codec is nil")
	}
	dataRaw, err := codec.Encode(data)
	if err != nil {
		return err
	}
	c.reqEntry = &message.Entry{
		ID:   id,
		Data: dataRaw,
	}
	return nil
}

// SetRequestMessage sets request message entry.
func (c *routeContext) SetRequestMessage(entry *message.Entry) Context {
	c.reqEntry = entry
	return c
}

// Bind implements Context.Bind method.
func (c *routeContext) Bind(v interface{}) error {
	if c.session.Codec() == nil {
		return fmt.Errorf("message codec is nil")
	}
	return c.session.Codec().Decode(c.reqEntry.Data, v)
}

// Response implements Context.Response method.
func (c *routeContext) Response() *message.Entry {
	return c.respEntry
}

// RawResponseData returns the not yet encoded response data.
func (c *routeContext) RawResponseData() interface{} {
	if c.rawRespData != nil {
		return c.rawRespData
	}
	if c.respEntry != nil {
		return c.respEntry.Data
	}
	return nil
}

// SetResponse implements Context.SetResponse method.
func (c *routeContext) SetResponse(id, data interface{}) error {
	codec := c.session.Codec()
	if codec == nil {
		return fmt.Errorf("codec is nil")
	}
	dataRaw, err := codec.Encode(data)
	if err != nil {
		return err
	}
	c.rawRespData = data
	c.respEntry = &message.Entry{
		ID:   id,
		Data: dataRaw,
	}
	return nil
}

// MustSetResponse implements Context.MustSetResponse method.
func (c *routeContext) MustSetResponse(id, data interface{}) Context {
	if err := c.SetResponse(id, data); err != nil {
		panic(err)
	}
	return c
}

// SetResponseMessage implements Context.SetResponseMessage method.
func (c *routeContext) SetResponseMessage(msg *message.Entry) Context {
	c.respEntry = msg
	return c
}

// Send implements Context.Send method.
func (c *routeContext) Send() {
	c.session.Send(c)
}

// SendTo implements Context.SendTo method.
func (c *routeContext) SendTo(sess Session) {
	sess.Send(c)
}

// Get implements Context.Get method.
func (c *routeContext) Get(key string) (value interface{}, exists bool) {
	c.mu.RLock()
	value, exists = c.storage[key]
	c.mu.RUnlock()
	return
}

// Set implements Context.Set method.
func (c *routeContext) Set(key string, value interface{}) {
	c.mu.Lock()
	if c.storage == nil {
		c.storage = make(map[string]interface{})
	}
	c.storage[key] = value
	c.mu.Unlock()
}

// Remove implements Context.Remove method.
func (c *routeContext) Remove(key string) {
	c.mu.Lock()
	delete(c.storage, key)
	c.mu.Unlock()
}

// Copy implements Context.Copy method.
func (c *routeContext) Copy() Context {
	return &routeContext{
		storage:   c.storage,
		session:   c.session,
		reqEntry:  c.reqEntry,
		respEntry: c.respEntry,
	}
}

func (c *routeContext) reset(sess *session, reqEntry *message.Entry) {
	c.session = sess
	c.reqEntry = reqEntry
	c.respEntry = nil
	c.storage = nil
}
