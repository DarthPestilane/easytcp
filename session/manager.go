package session

import (
	"github.com/DarthPestilane/easytcp/logger"
	"github.com/sirupsen/logrus"
	"sync"
)

var (
	managerOnce sync.Once
	manager     *Manager
)

// Manager 会话管理器，维护运行时产生的所有 Session
type Manager struct {
	// Sessions 所有 Session. key 是 Session.Id, value 是 *Session
	Sessions sync.Map
	log      *logrus.Entry
}

// Sessions 得到 Manager 实例
func Sessions() *Manager {
	managerOnce.Do(func() {
		manager = &Manager{
			log: logger.Default.WithField("scope", "session.Manager"),
		}
	})
	return manager
}

func (m *Manager) Add(s *Session) {
	if s == nil {
		return
	}
	m.Sessions.Store(s.Id, s)
}

func (m *Manager) Remove(id string) {
	m.Sessions.Delete(id)
}

func (m *Manager) Get(id string) *Session {
	sess, ok := m.Sessions.Load(id)
	if !ok {
		return nil
	}
	return sess.(*Session)
}

func (m *Manager) Range(fn func(id string, sess *Session) (next bool)) {
	m.Sessions.Range(func(key, value interface{}) bool {
		return fn(key.(string), value.(*Session))
	})
}
