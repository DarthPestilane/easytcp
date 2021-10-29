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
	Session() Session
	Message() *message.Entry
	Copy() Context
	SendTo(sess Session, id, data interface{}) error
	GetResponse() *message.Entry
	Bind(v interface{}) error
	MustBind(v interface{})
	Response(id, data interface{}) error
	MustDecodeTo(data []byte, v interface{})
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

// Get returns the value from c.storage by key.
func (c *routeContext) Get(key string) (value interface{}, exists bool) {
	c.mu.RLock()
	value, exists = c.storage[key]
	c.mu.RUnlock()
	return
}

// MustGet returns the value from c.storage by key.
// Panics if key does not exist.
func (c *routeContext) MustGet(key string) interface{} {
	if val, ok := c.Get(key); ok {
		return val
	}
	panic(fmt.Errorf("key `%s` does not exist", key))
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

// Message returns the request message entry.
func (c *routeContext) Message() *message.Entry {
	return c.reqEntry
}

// Bind binds the request message's raw data to v.
func (c *routeContext) Bind(v interface{}) error {
	if c.session.Codec() == nil {
		return fmt.Errorf("message codec is nil")
	}
	return c.session.Codec().Decode(c.reqEntry.Data, v)
}

// MustBind binds the request message's raw data to v.
// Panics if any error occurred.
func (c *routeContext) MustBind(v interface{}) {
	if err := c.Bind(v); err != nil {
		panic(err)
	}
}

// DecodeTo decodes data to v via codec.
func (c *routeContext) DecodeTo(data []byte, v interface{}) error {
	if c.session.Codec() == nil {
		return fmt.Errorf("message codec is nil")
	}
	return c.session.Codec().Decode(data, v)
}

// MustDecodeTo decodes data to v via codec.
// Panics if any error occurred.
func (c *routeContext) MustDecodeTo(data []byte, v interface{}) {
	if err := c.DecodeTo(data, v); err != nil {
		panic(err)
	}
}

// Encode encodes v using session's codec.
func (c *routeContext) Encode(v interface{}) ([]byte, error) {
	if c.session.Codec() == nil {
		return nil, fmt.Errorf("codec is not nil")
	}
	return c.session.Codec().Encode(v)
}

// MustEncode encodes v using session's codec.
// Panics if any error occurred.
func (c *routeContext) MustEncode(v interface{}) []byte {
	data, err := c.Encode(v)
	if err != nil {
		panic(err)
	}
	return data
}

// Remove deletes the key from storage.
func (c *routeContext) Remove(key string) {
	c.mu.Lock()
	delete(c.storage, key)
	c.mu.Unlock()
}

// session returns current session.
func (c *routeContext) Session() Session {
	return c.session
}

// SetResponse sets response entry with id and data.
func (c *routeContext) SetResponse(id interface{}, data []byte) {
	c.respEntry = &message.Entry{
		ID:   id,
		Data: data,
	}
}

// GetResponse returns response entry of context.
func (c *routeContext) GetResponse() *message.Entry {
	return c.respEntry
}

// Response creates and sets the response message to the context.
func (c *routeContext) Response(id, data interface{}) error {
	var dataRaw []byte
	if codec := c.session.Codec(); codec == nil {
		switch v := data.(type) {
		case []byte:
			dataRaw = v
		case *[]byte:
			dataRaw = *v
		case string:
			dataRaw = []byte(v)
		case *string:
			dataRaw = []byte(*v)
		case fmt.Stringer:
			dataRaw = []byte(v.String())
		default:
			return fmt.Errorf("data should be []byte, string or Stringer")
		}
	} else {
		var err error
		dataRaw, err = codec.Encode(data)
		if err != nil {
			return err
		}
	}
	c.SetResponse(id, dataRaw)
	return nil
}

// SendTo sends response message to the specified session.
// It should be called after Copy:
//   c.Copy().SendTo(...)
func (c *routeContext) SendTo(sess Session, id, data interface{}) error {
	if err := c.Response(id, data); err != nil {
		return err
	}
	return sess.Send(c)
}

// Send sends response message to current session.
// It should be called after Copy:
//   c.Copy().Send(...)
func (c *routeContext) Send(id, data interface{}) error {
	if err := c.Response(id, data); err != nil {
		return err
	}
	return c.session.Send(c)
}

// Copy returns a copy of the current context.
// This should be used when one wants to change the context after pushed to a channel.
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
