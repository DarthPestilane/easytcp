package easytcp

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestSessions(t *testing.T) {
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			Sessions()
			wg.Done()
		}()
	}
	wg.Wait()
	assert.NotNil(t, manager)
	assert.Equal(t, manager, Sessions())
}

func TestManager_Add(t *testing.T) {
	mg := &Manager{}
	assert.NotPanics(t, func() { mg.Add(nil) })

	sess := NewSession(nil, &SessionOption{})

	mg.Add(sess)

	v, ok := mg.Sessions.Load(sess.ID())
	assert.True(t, ok)
	assert.Equal(t, v, sess)
}

func TestManager_Get(t *testing.T) {
	mg := &Manager{}
	assert.Nil(t, mg.Get("not found"))

	sess := NewSession(nil, &SessionOption{})

	mg.Sessions.Store(sess.ID(), sess)
	s := mg.Get(sess.ID())
	assert.NotNil(t, s)
	assert.Equal(t, s, sess)
}

func TestManager_Range(t *testing.T) {
	mg := &Manager{}
	var count int
	mg.Range(func(id string, sess *Session) (next bool) {
		count++
		return true
	})
	assert.Zero(t, count)

	sess := NewSession(nil, &SessionOption{})

	mg.Add(sess)
	count = 0
	mg.Range(func(id string, s *Session) (next bool) {
		assert.Equal(t, sess.ID(), id)
		assert.Equal(t, s, sess)
		count++
		return true
	})
	assert.Equal(t, count, 1)
}

func TestManager_Remove(t *testing.T) {
	mg := &Manager{}
	assert.NotPanics(t, func() {
		mg.Remove("not found")
	})
	mg.Sessions.Store("test", "test")
	mg.Remove("test")
	_, found := mg.Sessions.Load("test")
	assert.False(t, found)
}
