package easytcp

import (
	"sync"
)

var (
	managerOnce sync.Once
	manager     *SessionManager
)

// SessionManager manages all the sessions in application runtime.
type SessionManager struct {
	// Sessions keeps all sessions.
	// Key is session's ID, value is *Session
	Sessions sync.Map
}

// Sessions returns a SessionManager pointer in a singleton way.
func Sessions() *SessionManager {
	managerOnce.Do(func() {
		manager = &SessionManager{}
	})
	return manager
}

// Add adds a session to Sessions.
// If the ID of s already existed in Sessions, it replaces the value with the s.
func (m *SessionManager) Add(s *Session) {
	if s == nil {
		return
	}
	m.Sessions.Store(s.ID(), s)
}

// Remove removes a session from Sessions.
// Parameter id should be the session's id.
func (m *SessionManager) Remove(id string) {
	m.Sessions.Delete(id)
}

// Get returns a Session when found by the id,
// returns nil otherwise.
func (m *SessionManager) Get(id string) *Session {
	sess, ok := m.Sessions.Load(id)
	if !ok {
		return nil
	}
	return sess.(*Session)
}

// Range calls fn sequentially for each id and sess present in the Sessions.
// If fn returns false, range stops the iteration.
func (m *SessionManager) Range(fn func(id string, sess *Session) (next bool)) {
	m.Sessions.Range(func(key, value interface{}) bool {
		return fn(key.(string), value.(*Session))
	})
}
