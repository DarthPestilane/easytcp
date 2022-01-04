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
	return &routeContext{
		rawCtx: context.Background(),
	}
}

// Context is a generic context in a message routing.
// It allows us to pass variables between handler and middlewares.
type Context interface {
	context.Context

	// WithContext sets the underline context.
	// It's very useful to control the workflow when send to response channel.
	WithContext(ctx context.Context) Context

	// Session returns the current session.
	Session() Session

	// SetSession sets session.
	SetSession(sess Session) Context

	// Request returns request message entry.
	Request() *message.Entry

	// SetRequest encodes data with session's codec and sets request message entry.
	SetRequest(id, data interface{}) error

	// MustSetRequest encodes data with session's codec and sets request message entry.
	// panics on error.
	MustSetRequest(id, data interface{}) Context

	// SetRequestMessage sets request message entry directly.
	SetRequestMessage(entry *message.Entry) Context

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
	rawCtx    context.Context
	mu        sync.RWMutex
	storage   map[string]interface{}
	session   Session
	reqEntry  *message.Entry
	respEntry *message.Entry
}

// Deadline implements the context.Context Deadline method.
func (c *routeContext) Deadline() (time.Time, bool) {
	return c.rawCtx.Deadline()
}

// Done implements the context.Context Done method.
func (c *routeContext) Done() <-chan struct{} {
	return c.rawCtx.Done()
}

// Err implements the context.Context Err method.
func (c *routeContext) Err() error {
	return c.rawCtx.Err()
}

// Value implements the context.Context Value method.
func (c *routeContext) Value(key interface{}) interface{} {
	if keyAsString, ok := key.(string); ok {
		val, _ := c.Get(keyAsString)
		return val
	}
	return nil
}

// WithContext sets the underline context.
func (c *routeContext) WithContext(ctx context.Context) Context {
	c.rawCtx = ctx
	return c
}

// Session implements Context.Session method.
func (c *routeContext) Session() Session {
	return c.session
}

// SetSession sets session.
func (c *routeContext) SetSession(sess Session) Context {
	c.session = sess
	return c
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

// MustSetRequest implements Context.MustSetRequest method.
func (c *routeContext) MustSetRequest(id, data interface{}) Context {
	if err := c.SetRequest(id, data); err != nil {
		panic(err)
	}
	return c
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
		rawCtx:    c.rawCtx,
		storage:   c.storage,
		session:   c.session,
		reqEntry:  c.reqEntry,
		respEntry: c.respEntry,
	}
}

func (c *routeContext) reset(sess *session, reqEntry *message.Entry) {
	c.rawCtx = context.Background()
	c.session = sess
	c.reqEntry = reqEntry
	c.respEntry = nil
	c.storage = nil
}
