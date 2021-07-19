package easytcp

import (
	"fmt"
	"net"
	"time"
)

//go:generate mockgen -destination mock/net_mock.go -package mock net Listener,Error

// Server is a server for TCP connections.
type Server struct {
	Listener net.Listener

	// Packer is the message packer, will be passed to session.
	Packer Packer

	// Codec is the message codec, will be passed to session.
	Codec Codec

	// OnSessionCreate is a event hook, will be invoked when session's created.
	OnSessionCreate func(sess *Session)

	// OnSessionClose is a event hook, will be invoked when session's closed.
	OnSessionClose func(sess *Session)

	socketReadBufferSize  int
	socketWriteBufferSize int
	socketSendDelay       bool

	writeBufferSize int
	readBufferSize  int
	readTimeout     time.Duration
	writeTimeout    time.Duration
	printRoutes     bool
	router          *Router
	accepting       chan struct{}
	stopped         chan struct{}
}

// ServerOption is the option for Server.
type ServerOption struct {
	SocketReadBufferSize  int           // sets the socket read buffer size.
	SocketWriteBufferSize int           // sets the socket write buffer size.
	SocketSendDelay       bool          // sets the socket delay or not.
	ReadTimeout           time.Duration // sets the timeout for connection read.
	WriteTimeout          time.Duration // sets the timeout for connection write.
	Packer                Packer        // packs and unpacks packet payload, default packer is the packet.DefaultPacker.
	Codec                 Codec         // encodes and decodes the message data, can be nil
	WriteBufferSize       int           // sets the write channel buffer size, 1024 will be used if < 0.
	ReadBufferSize        int           // sets the read channel buffer size, 1024 will be used if < 0.
	DontPrintRoutes       bool          // whether to print registered route handlers to the console.
}

// NewServer creates a Server pointer according to opt.
func NewServer(opt *ServerOption) *Server {
	if opt.Packer == nil {
		opt.Packer = NewDefaultPacker()
	}
	if opt.WriteBufferSize < 0 {
		opt.WriteBufferSize = 1024
	}
	if opt.ReadBufferSize < 0 {
		opt.ReadBufferSize = 1024
	}
	return &Server{
		socketReadBufferSize:  opt.SocketReadBufferSize,
		socketWriteBufferSize: opt.SocketWriteBufferSize,
		socketSendDelay:       opt.SocketSendDelay,
		writeBufferSize:       opt.WriteBufferSize,
		readBufferSize:        opt.ReadBufferSize,
		readTimeout:           opt.ReadTimeout,
		writeTimeout:          opt.WriteTimeout,
		Packer:                opt.Packer,
		Codec:                 opt.Codec,
		printRoutes:           !opt.DontPrintRoutes,
		router:                newRouter(),
		accepting:             make(chan struct{}),
		stopped:               make(chan struct{}),
	}
}

// Serve starts to listen TCP and keep accepting TCP connection in a loop.
// Accepting loop will break when error occurred, and the error will be returned.
func (s *Server) Serve(addr string) error {
	address, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return err
	}
	lis, err := net.ListenTCP("tcp", address)
	if err != nil {
		return err
	}
	s.Listener = lis
	if s.printRoutes {
		s.router.printHandlers(fmt.Sprintf("tcp://%s", s.Listener.Addr()))
	}
	return s.acceptLoop()
}

// acceptLoop accepts TCP connections in a loop, and handle connections in goroutines.
// Returns error when error occurred.
func (s *Server) acceptLoop() error {
	close(s.accepting)
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			select {
			case <-s.stopped:
				return ErrServerStopped
			default:
			}
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				tempDelay := time.Millisecond * 5
				Log.Tracef("accept err: %s; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return fmt.Errorf("accept err: %s", err)
		}
		if s.socketReadBufferSize > 0 {
			if err := conn.(*net.TCPConn).SetReadBuffer(s.socketReadBufferSize); err != nil {
				return fmt.Errorf("conn set read buffer err: %s", err)
			}
		}
		if s.socketWriteBufferSize > 0 {
			if err := conn.(*net.TCPConn).SetWriteBuffer(s.socketWriteBufferSize); err != nil {
				return fmt.Errorf("conn set write buffer err: %s", err)
			}
		}
		if s.socketSendDelay {
			if err := conn.(*net.TCPConn).SetNoDelay(false); err != nil {
				return fmt.Errorf("conn set no delay err: %s", err)
			}
		}
		go s.handleConn(conn)
	}
}

// handleConn creates a new session according to conn,
// handles the message through the session in different goroutines,
// and waits until the session's closed.
func (s *Server) handleConn(conn net.Conn) {
	sess := newSession(conn, &SessionOption{
		Packer:          s.Packer,
		Codec:           s.Codec,
		ReadBufferSize:  s.readBufferSize,
		WriteBufferSize: s.writeBufferSize,
	})
	Sessions().Add(sess)
	if s.OnSessionCreate != nil {
		go s.OnSessionCreate(sess)
	}
	go s.router.routeLoop(sess)
	go sess.readLoop(s.readTimeout)
	go sess.writeLoop(s.writeTimeout)
	<-sess.closed
	Sessions().Remove(sess.ID()) // session has been closed, remove it
	if s.OnSessionClose != nil {
		go s.OnSessionClose(sess)
	}
	if err := conn.Close(); err != nil {
		Log.Tracef("connection close err: %s", err)
	}
}

// Stop stops server by closing all the TCP sessions and the listener.
func (s *Server) Stop() error {
	closedNum := 0
	Sessions().Range(func(id string, sess *Session) (next bool) {
		sess.Close()
		closedNum++
		return true
	})
	Log.Tracef("%d session(s) closed", closedNum)
	close(s.stopped)
	return s.Listener.Close()
}

// AddRoute registers message handler and middlewares to the router.
func (s *Server) AddRoute(msgID interface{}, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	s.router.register(msgID, handler, middlewares...)
}

// Use registers global middlewares to the router.
func (s *Server) Use(middlewares ...MiddlewareFunc) {
	s.router.registerMiddleware(middlewares...)
}

// NotFoundHandler sets the not-found handler for router.
func (s *Server) NotFoundHandler(handler HandlerFunc) {
	s.router.setNotFoundHandler(handler)
}
