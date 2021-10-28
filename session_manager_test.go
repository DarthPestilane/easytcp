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

func TestManager_AddGetAndRemove(t *testing.T) {
	mg := &SessionManager{}

	// should not add nil
	assert.NotPanics(t, func() { mg.Add(nil) })

	assert.Nil(t, mg.Get("not found"))

	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sess := newSession(nil, &sessionOption{})
			mg.Add(sess)
			s := mg.Get(sess.id)
			assert.NotNil(t, s)
			assert.Equal(t, s, sess)
			mg.Remove(sess.id)
			assert.Nil(t, mg.Get(sess.id))
		}()
	}
	wg.Wait()
}

func TestManager_Range(t *testing.T) {
	mg := &SessionManager{}
	var count int
	mg.Range(func(id string, sess ISession) (next bool) {
		count++
		return true
	})
	assert.Zero(t, count)

	sess := newSession(nil, &sessionOption{})
	sess2 := newSession(nil, &sessionOption{})
	mg.Add(sess)
	mg.Add(sess2)

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		count := 0
		mg.Range(func(id string, s ISession) (next bool) {
			count++
			return false
		})
		assert.Equal(t, count, 1)
	}()

	go func() {
		defer wg.Done()
		count := 0
		mg.Range(func(id string, s ISession) (next bool) {
			count++
			return true
		})
		assert.Equal(t, count, 2)
	}()
	wg.Wait()
}
