package core

import (
	"fmt"
	"net"
	"sync"
)

type Server struct {
	// Addr the address: 127.0.0.1
	Addr string
	// Port eg: 8765
	Port int

	mu sync.Mutex

	listener net.Listener

	// route handlers
	// key is the route path
	handler map[string]HandlerFunc

	// hook functions
	onConnectedFn  ConnectHookFunc
	onDisconnectFn ConnectHookFunc

	bufferSize int
}

type HandlerFunc func(ctx *Context)
type ConnectHookFunc func(conn *Connection)

func NewServer(addr string, port int) *Server {
	return &Server{
		Addr:       addr,
		Port:       port,
		handler:    make(map[string]HandlerFunc),
		bufferSize: 1024,
	}
}

func (s *Server) Serve() error {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Addr, s.Port))
	if err != nil {
		return fmt.Errorf("listen tcp failed: %w", err)
	}
	s.listener = lis
	return s.keepAccepting()
}

func (s *Server) SetBufferSize(n int) {
	if n > 0 {
		s.bufferSize = n
	}
}

func (s *Server) OnConnected(fn ConnectHookFunc) {
	s.onConnectedFn = fn
}

func (s *Server) OnDisconnect(fn ConnectHookFunc) {
	s.onDisconnectFn = fn
}

func (s *Server) AddRoute(routePath string, fn HandlerFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handler[routePath] = fn
}

func (s *Server) keepAccepting() error {
	for {
		rawConn, err := s.listener.Accept()
		if err != nil {
			return err
		}

		// 拿到连接后，放到goroutine里处理，然后接着拿下一个连接
		go func() {
			conn := NewConnection(rawConn, ConnectionOption{
				BufferSize: s.bufferSize,
				Handler:    s.handler,
			})

			go conn.KeepReading()
			go conn.KeepWriting()

			if s.onConnectedFn != nil {
				s.onConnectedFn(conn)
			}

			<-conn.Closed

			if s.onDisconnectFn != nil {
				s.onDisconnectFn(conn)
			}
		}()
	}
}
