package easytcp

import (
	"fmt"
	"sync"
)

// NewMessage creates a Message pointer.
func NewMessage(id interface{}, data []byte) *Message {
	return &Message{
		id:   id,
		data: data,
	}
}

// Message is the abstract of inbound and outbound message.
type Message struct {
	id      interface{}
	data    []byte
	storage map[string]interface{}
	mu      sync.RWMutex
}

// ID returns the id of current message.
func (m *Message) ID() interface{} {
	return m.id
}

// Data returns the data part of current message.
func (m *Message) Data() []byte {
	return m.data
}

// Set stores kv pair.
func (m *Message) Set(key string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.storage == nil {
		m.storage = make(map[string]interface{})
	}
	m.storage[key] = value
}

// Get retrieves the value according to the key.
func (m *Message) Get(key string) (value interface{}, exists bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, exists = m.storage[key]
	return
}

// MustGet retrieves the value according to the key.
// Panics if key does not exist.
func (m *Message) MustGet(key string) interface{} {
	if v, ok := m.Get(key); ok {
		return v
	}
	panic(fmt.Errorf("key `%s` does not exist", key))
}

// Remove deletes the key from storage.
func (m *Message) Remove(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.storage, key)
}

// Reset resets m.
func (m *Message) Reset(id interface{}, data []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.id = id
	m.data = data
	m.storage = nil
}
