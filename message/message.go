package message

import (
	"sync"
)

// Entry is the unpacked message object.
type Entry struct {
	ID      interface{}
	Data    []byte
	storage sync.Map
}

func (e *Entry) Set(key string, value interface{}) {
	e.storage.Store(key, value)
}

func (e *Entry) Get(key string) (interface{}, bool) {
	return e.storage.Load(key)
}
