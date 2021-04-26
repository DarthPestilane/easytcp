package tcp_demo

import (
	"bufio"
	"demo/tcp_demo/util"
	"demo/tcp_demo/util/message"
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"sync"
)

type Server struct {
	Addr    string
	Port    int
	mu      sync.Mutex
	handler map[string]HandlerFunc
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
	return s.accept(lis)
}

func (s *Server) AddRoute(routePath string, fn HandlerFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handler[routePath] = fn
}

func (s *Server) handleMessage(conn net.Conn) {
	for {
		connReader := bufio.NewReader(conn)
		head, err := connReader.ReadBytes('|')
		if err != nil {
			if util.IsEOF(err) {
				logrus.Infof("disconnected!!! %s: %s", conn.RemoteAddr(), err)
				return
			}
			logrus.Errorf("whoops...read conn %s head failed: %s", conn.RemoteAddr(), err)
			continue
		}
		headStruct, err := message.ExtractHead(head)
		if err != nil {
			logrus.Error(err)
			continue
		}
		body := make([]byte, headStruct.Length)
		n, err := connReader.Read(body)
		if err != nil {
			if util.IsEOF(err) {
				logrus.Infof("disconnected!!! %s: %s", conn.RemoteAddr(), err)
				return
			}
			logrus.Errorf("whoops...read conn %s body failed: %s", conn.RemoteAddr(), err)
			continue
		}
		ctx := NewContext()
		ctx.setConn(conn).setBody(body[:n])
		s.handler[headStruct.RoutePath](ctx)
	}
}

func (s *Server) accept(lis net.Listener) error {
	for {
		conn, err := lis.Accept()
		if err != nil {
			return err
		}
		go s.handleMessage(conn)
	}
}
