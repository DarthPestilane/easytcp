package easytcp

import (
	"context"
	"fmt"
	"github.com/DarthPestilane/easytcp/message"
	"sync"
	"time"
)

// Context is a generic context in a message routing.
// It allows us to pass variables between handler and middlewares.
type Context interface {
	context.Context

	// Session returns the current session.
	Session() Session

	// Request returns request message entry.
	Request() *message.Entry

	// Bind decodes request message entry to v.
	Bind(v interface{}) error

	// Response returns the response message entry.
	Response() *message.Entry

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
	mu        sync.RWMutex
	storage   map[string]interface{}
	session   Session
	reqEntry  *message.Entry
	respEntry *message.Entry
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

func (c *routeContext) Session() Session {
	return c.session
}

func (c *routeContext) Request() *message.Entry {
	return c.reqEntry
}

func (c *routeContext) Bind(v interface{}) error {
	if c.session.Codec() == nil {
		return fmt.Errorf("message codec is nil")
	}
	return c.session.Codec().Decode(c.reqEntry.Data, v)
}

func (c *routeContext) Response() *message.Entry {
	return c.respEntry
}

func (c *routeContext) SetResponse(id, data interface{}) error {
	if c.Session().Codec() == nil {
		return fmt.Errorf("codec is nil")
	}
	dataRaw, err := c.Session().Codec().Encode(data)
	if err != nil {
		return err
	}
	c.respEntry = &message.Entry{
		ID:   id,
		Data: dataRaw,
	}
	return nil
}

func (c *routeContext) MustSetResponse(id, data interface{}) Context {
	if err := c.SetResponse(id, data); err != nil {
		panic(err)
	}
	return c
}

func (c *routeContext) SetResponseMessage(msg *message.Entry) Context {
	c.respEntry = msg
	return c
}

func (c *routeContext) Send() {
	if c.respEntry == nil {
		return
	}
	c.session.Send(c)
}

func (c *routeContext) SendTo(sess Session) {
	if c.respEntry == nil {
		return
	}
	sess.Send(c)
}

// Get returns the value from c.storage by key.
func (c *routeContext) Get(key string) (value interface{}, exists bool) {
	c.mu.RLock()
	value, exists = c.storage[key]
	c.mu.RUnlock()
	return
}

// Set sets the value in c.storage.
func (c *routeContext) Set(key string, value interface{}) {
	c.mu.Lock()
	if c.storage == nil {
		c.storage = make(map[string]interface{})
	}
	c.storage[key] = value
	c.mu.Unlock()
}

func (c *routeContext) Remove(key string) {
	c.mu.Lock()
	delete(c.storage, key)
	c.mu.Unlock()
}

func (c *routeContext) Copy() Context {
	cp := routeContext{
		storage:   c.storage,
		session:   c.session,
		reqEntry:  c.reqEntry,
		respEntry: c.respEntry,
	}
	return &cp
}

func (c *routeContext) reset(sess *session, reqEntry *message.Entry) {
	c.session = sess
	c.reqEntry = reqEntry
	c.respEntry = nil
	c.storage = nil
}
