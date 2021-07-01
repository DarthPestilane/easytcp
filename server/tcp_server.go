package server

import (
	"fmt"
	"github.com/DarthPestilane/easytcp/logger"
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/DarthPestilane/easytcp/router"
	"github.com/DarthPestilane/easytcp/session"
	"net"
	"time"
)

// TCPServer is a server for TCP connections.
// TCPServer implements the Server interface.
type TCPServer struct {
	socketRWBufferSize int
	writeBufferSize    int
	readBufferSize     int
	readTimeout        time.Duration
	writeTimeout       time.Duration
	printRoutes        bool
	listener           net.Listener
	msgPacker          packet.Packer
	msgCodec           packet.Codec
	router             *router.Router
	accepting          chan struct{}
	stopped            chan struct{}
}

var _ Server = &TCPServer{}

// TCPOption is the option for TCPServer.
type TCPOption struct {
	SocketRWBufferSize int           // sets the socket read write buffer
	ReadTimeout        time.Duration // sets the timeout for connection read
	WriteTimeout       time.Duration // sets the timeout for connection write
	MsgPacker          packet.Packer // packs and unpacks packet payload, default packer is the packet.DefaultPacker.
	MsgCodec           packet.Codec  // encodes and decodes the message data, can be nil
	WriteBufferSize    int           // sets the write channel buffer size, 1024 will be used if < 0.
	ReadBufferSize     int           // sets the read channel buffer size, 1024 will be used if < 0.
	DontPrintRoutes    bool          // whether to print registered route handlers to the console.
}

// NewTCPServer creates a TCPServer pointer according to opt.
func NewTCPServer(opt *TCPOption) *TCPServer {
	if opt.MsgPacker == nil {
		opt.MsgPacker = &packet.DefaultPacker{}
	}
	if opt.WriteBufferSize < 0 {
		opt.WriteBufferSize = 1024
	}
	if opt.ReadBufferSize < 0 {
		opt.ReadBufferSize = 1024
	}
	return &TCPServer{
		socketRWBufferSize: opt.SocketRWBufferSize,
		writeBufferSize:    opt.WriteBufferSize,
		readBufferSize:     opt.ReadBufferSize,
		readTimeout:        opt.ReadTimeout,
		writeTimeout:       opt.WriteTimeout,
		msgPacker:          opt.MsgPacker,
		msgCodec:           opt.MsgCodec,
		printRoutes:        !opt.DontPrintRoutes,
		router:             router.NewRouter(),
		accepting:          make(chan struct{}),
		stopped:            make(chan struct{}),
	}
}

// Serve implements the Server Serve method.
// Serve starts to listen TCP and keep accepting TCP connection in a loop.
// Accepting loop will break when error occurred, and the error will be returned.
func (s *TCPServer) Serve(addr string) error {
	address, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return err
	}
	lis, err := net.ListenTCP("tcp", address)
	if err != nil {
		return err
	}
	s.listener = lis
	if s.printRoutes {
		s.router.PrintHandlers(fmt.Sprintf("tcp://%s", s.listener.Addr()))
	}
	return s.acceptLoop()
}

// acceptLoop accepts TCP connections in a loop, and handle connections in goroutines.
// Returns error when error occurred.
func (s *TCPServer) acceptLoop() error {
	close(s.accepting)
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if isStopped(s.stopped) {
				return ErrServerStopped
			}
			if isTempErr(err) {
				tempDelay := time.Millisecond * 5
				logger.Log.Tracef("accept err: %s; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return fmt.Errorf("accept err: %s", err)
		}
		if s.socketRWBufferSize > 0 {
			if err := conn.(*net.TCPConn).SetReadBuffer(s.socketRWBufferSize); err != nil {
				return fmt.Errorf("conn set read buffer err: %s", err)
			}
			if err := conn.(*net.TCPConn).SetWriteBuffer(s.socketRWBufferSize); err != nil {
				return fmt.Errorf("conn set write buffer err: %s", err)
			}
		}
		go s.handleConn(conn)
	}
}

// handleConn creates a new session according to conn,
// handles the message through the session in different goroutines,
// and waits until the session's closed.
func (s *TCPServer) handleConn(conn net.Conn) {
	sess := session.NewTCPSession(conn, &session.TCPSessionOption{
		Packer:          s.msgPacker,
		Codec:           s.msgCodec,
		ReadBufferSize:  s.readBufferSize,
		WriteBufferSize: s.writeBufferSize,
	})
	session.Sessions().Add(sess)
	go s.router.RouteLoop(sess)
	go sess.ReadLoop(s.readTimeout)
	go sess.WriteLoop(s.writeTimeout)
	sess.WaitUntilClosed()
	session.Sessions().Remove(sess.ID()) // session has been closed, remove it
	logger.Log.Tracef("session closed")
	if err := conn.Close(); err != nil {
		logger.Log.Tracef("connection close err: %s", err)
	}
}

// Stop implements the Server Stop method.
// Stop stops server by closing all the TCP sessions and the listener.
func (s *TCPServer) Stop() error {
	closedNum := 0
	session.Sessions().Range(func(id string, sess session.Session) (next bool) {
		if tcpSess, ok := sess.(*session.TCPSession); ok {
			tcpSess.Close()
			closedNum++
		}
		return true
	})
	logger.Log.Tracef("%d session(s) closed", closedNum)
	close(s.stopped)
	return s.listener.Close()
}

// AddRoute implements the Server AddRoute method.
// AddRoute registers message handler and middlewares to the router.
func (s *TCPServer) AddRoute(msgID uint, handler router.HandlerFunc, middlewares ...router.MiddlewareFunc) {
	s.router.Register(msgID, handler, middlewares...)
}

// Use implements the Server Use method.
// Use registers global middlewares to the router.
func (s *TCPServer) Use(middlewares ...router.MiddlewareFunc) {
	s.router.RegisterMiddleware(middlewares...)
}
