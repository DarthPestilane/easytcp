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
	mu       sync.RWMutex
	sessions map[string]Session
}

// Sessions returns a SessionManager pointer in a singleton way.
func Sessions() *SessionManager {
	managerOnce.Do(func() {
		manager = &SessionManager{}
	})
	return manager
}

// Add adds a session to sessions.
// If the ID of s already existed in sessions, it replaces the value with the s.
func (m *SessionManager) Add(s Session) {
	if s == nil {
		return
	}
	m.mu.Lock()
	if m.sessions == nil {
		m.sessions = make(map[string]Session)
	}
	m.sessions[s.ID()] = s
	m.mu.Unlock()
}

// Remove removes a session from sessions.
// Parameter id should be the session's id.
func (m *SessionManager) Remove(id string) {
	m.mu.Lock()
	delete(m.sessions, id)
	m.mu.Unlock()
}

// Get returns a session when found by the id,
// returns nil otherwise.
func (m *SessionManager) Get(id string) Session {
	m.mu.RLock()
	sess := m.sessions[id]
	m.mu.RUnlock()
	return sess
}

// Range calls fn sequentially for each id and sess present in the sessions.
// If fn returns false, range stops the iteration.
func (m *SessionManager) Range(fn func(id string, sess Session) (next bool)) {
	m.mu.RLock()
	for id, sess := range m.sessions {
		if !fn(id, sess) {
			break
		}
	}
	m.mu.RUnlock()
}
