package easytcp

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"sync"
)

type Server struct {
	Addr     string
	Port     int
	mu       sync.Mutex
	handler  map[string]HandlerFunc
	listener net.Listener
}

type HandlerFunc func(ctx *Context)

func NewServer(addr string, port int) *Server {
	return &Server{
		Addr:    addr,
		Port:    port,
		handler: make(map[string]HandlerFunc),
	}
}

func (s *Server) Serve() error {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Addr, s.Port))
	if err != nil {
		return fmt.Errorf("listen tcp failed: %w", err)
	}
	s.listener = lis
	return s.accept()
}

func (s *Server) AddRoute(routePath string, fn HandlerFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handler[routePath] = fn
}

func (s *Server) accept() error {
	for {
		rawConn, err := s.listener.Accept()
		if err != nil {
			return err
		}
		conn := NewConnection(rawConn, s)
		logrus.Infof("connected!! %s", conn.RemoteAddr())
		go conn.KeepReading()
		go conn.KeepWriting()
	}
}
