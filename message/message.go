package message

import (
	"fmt"
	"sync"
)

// Entry is the unpacked message object.
type Entry struct {
	ID      interface{}
	Data    []byte
	storage map[string]interface{}
	mu      sync.RWMutex
}

// Set stores kv pair.
func (e *Entry) Set(key string, value interface{}) {
	e.mu.Lock()
	if e.storage == nil {
		e.storage = make(map[string]interface{})
	}
	e.storage[key] = value
	e.mu.Unlock()
}

// Get retrieves the value according to the key.
func (e *Entry) Get(key string) (value interface{}, exists bool) {
	e.mu.RLock()
	value, exists = e.storage[key]
	e.mu.RUnlock()
	return
}

// MustGet retrieves the value according to the key.
// Panics if key does not exist.
func (e *Entry) MustGet(key string) interface{} {
	if v, ok := e.Get(key); ok {
		return v
	}
	panic(fmt.Errorf("key `%s` does not exist", key))
}
