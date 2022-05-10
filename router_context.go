package easytcp

import (
	"context"
	"fmt"
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

	// Request returns request message.
	Request() *Message

	// SetRequest encodes data with session's codec and sets request message.
	SetRequest(id, data interface{}) error

	// MustSetRequest encodes data with session's codec and sets request message.
	// panics on error.
	MustSetRequest(id, data interface{}) Context

	// SetRequestMessage sets request message directly.
	SetRequestMessage(msg *Message) Context

	// Bind decodes request message to v.
	Bind(v interface{}) error

	// Response returns the response message.
	Response() *Message

	// SetResponse encodes data with session's codec and sets response message.
	SetResponse(id, data interface{}) error

	// MustSetResponse encodes data with session's codec and sets response message.
	// panics on error.
	MustSetResponse(id, data interface{}) Context

	// SetResponseMessage sets response message directly.
	SetResponseMessage(msg *Message) Context

	// Send sends itself to current session.
	Send() bool

	// SendTo sends itself to session.
	SendTo(session Session) bool

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
	rawCtx  context.Context
	mu      sync.RWMutex
	storage map[string]interface{}
	session Session
	reqMsg  *Message
	respMsg *Message
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
func (c *routeContext) Request() *Message {
	return c.reqMsg
}

// SetRequest sets request by id and data.
func (c *routeContext) SetRequest(id, data interface{}) error {
	codec := c.session.Codec()
	if codec == nil {
		return fmt.Errorf("codec is nil")
	}
	dataBytes, err := codec.Encode(data)
	if err != nil {
		return err
	}
	c.reqMsg = NewMessage(id, dataBytes)
	return nil
}

// MustSetRequest implements Context.MustSetRequest method.
func (c *routeContext) MustSetRequest(id, data interface{}) Context {
	if err := c.SetRequest(id, data); err != nil {
		panic(err)
	}
	return c
}

// SetRequestMessage sets request message.
func (c *routeContext) SetRequestMessage(msg *Message) Context {
	c.reqMsg = msg
	return c
}

// Bind implements Context.Bind method.
func (c *routeContext) Bind(v interface{}) error {
	if c.session.Codec() == nil {
		return fmt.Errorf("message codec is nil")
	}
	return c.session.Codec().Decode(c.reqMsg.Data(), v)
}

// Response implements Context.Response method.
func (c *routeContext) Response() *Message {
	return c.respMsg
}

// SetResponse implements Context.SetResponse method.
func (c *routeContext) SetResponse(id, data interface{}) error {
	codec := c.session.Codec()
	if codec == nil {
		return fmt.Errorf("codec is nil")
	}
	dataBytes, err := codec.Encode(data)
	if err != nil {
		return err
	}
	c.respMsg = NewMessage(id, dataBytes)
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
func (c *routeContext) SetResponseMessage(msg *Message) Context {
	c.respMsg = msg
	return c
}

// Send implements Context.Send method.
func (c *routeContext) Send() bool {
	return c.session.Send(c)
}

// SendTo implements Context.SendTo method.
func (c *routeContext) SendTo(sess Session) bool {
	return sess.Send(c)
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
		rawCtx:  c.rawCtx,
		storage: c.storage,
		session: c.session,
		reqMsg:  c.reqMsg,
		respMsg: c.respMsg,
	}
}

func (c *routeContext) reset() {
	c.rawCtx = context.Background()
	c.session = nil
	c.reqMsg = nil
	c.respMsg = nil
	c.storage = nil
}
