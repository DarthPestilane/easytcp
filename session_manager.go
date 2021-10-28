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
	sessions map[string]ISession
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
func (m *SessionManager) Add(s ISession) {
	if s == nil {
		return
	}
	m.mu.Lock()
	if m.sessions == nil {
		m.sessions = make(map[string]ISession)
	}
	m.sessions[s.ID()] = s
	m.mu.Unlock()
}

// Remove removes a session from Sessions.
// Parameter id should be the session's id.
func (m *SessionManager) Remove(id string) {
	m.mu.Lock()
	delete(m.sessions, id)
	m.mu.Unlock()
}

// Get returns a Session when found by the id,
// returns nil otherwise.
func (m *SessionManager) Get(id string) ISession {
	m.mu.RLock()
	sess := m.sessions[id]
	m.mu.RUnlock()
	return sess
}

// Range calls fn sequentially for each id and sess present in the Sessions.
// If fn returns false, range stops the iteration.
func (m *SessionManager) Range(fn func(id string, sess ISession) (next bool)) {
	m.mu.RLock()
	for id, sess := range m.sessions {
		if !fn(id, sess) {
			break
		}
	}
	m.mu.RUnlock()
}
