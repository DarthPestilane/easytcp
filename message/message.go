package message

import (
	"fmt"
	"sync"
)

// Entry is the unpacked message object.
type Entry struct {
	ID      interface{}
	Data    []byte
	storage sync.Map
}

// Set stores kv pair.
func (e *Entry) Set(key string, value interface{}) {
	e.storage.Store(key, value)
}

// Get retrieves the value according to the key.
func (e *Entry) Get(key string) (interface{}, bool) {
	return e.storage.Load(key)
}

// MustGet retrieves the value according to the key.
// Panics if key does not exist.
func (e *Entry) MustGet(key string) interface{} {
	if v, ok := e.Get(key); ok {
		return v
	}
	panic(fmt.Errorf("key `%s` does not exist", key))
}
