package server

import (
	"fmt"
	"github.com/DarthPestilane/easytcp/logger"
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/DarthPestilane/easytcp/router"
	"github.com/DarthPestilane/easytcp/session"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

// TCPServer is a server for TCP connections.
// TCPServer implements the Server interface.
type TCPServer struct {
	rwBufferSize int
	readTimeout  time.Duration
	writeTimeout time.Duration
	listener     *net.TCPListener
	log          *logrus.Entry
	msgPacker    packet.Packer
	msgCodec     packet.Codec
	router       *router.Router
	accepting    chan struct{}
	stopped      chan struct{}
}

var _ Server = &TCPServer{}

// TCPOption is the option for TCPServer.
type TCPOption struct {
	RWBufferSize int           // RWBufferSize is socket read write buffer
	ReadTimeout  time.Duration // sets the timeout for connection read
	WriteTimeout time.Duration // sets the timeout for connection write
	MsgPacker    packet.Packer // packs and unpacks the message packet
	MsgCodec     packet.Codec  // encodes and decodes the message data
}

// NewTCPServer creates a TCPServer pointer according to opt.
func NewTCPServer(opt TCPOption) *TCPServer {
	if opt.MsgPacker == nil {
		opt.MsgPacker = &packet.DefaultPacker{}
	}
	return &TCPServer{
		log:          logger.Default.WithField("scope", "server.TCPServer"),
		rwBufferSize: opt.RWBufferSize,
		readTimeout:  opt.ReadTimeout,
		writeTimeout: opt.WriteTimeout,
		msgPacker:    opt.MsgPacker,
		msgCodec:     opt.MsgCodec,
		router:       router.NewRouter(),
		accepting:    make(chan struct{}),
		stopped:      make(chan struct{}),
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

	return s.acceptLoop()
}

// acceptLoop accepts TCP connections in a loop, and handle connections in goroutines.
// Returns error when error occurred.
func (s *TCPServer) acceptLoop() error {
	close(s.accepting)
	for {
		conn, err := s.listener.AcceptTCP()
		if err != nil {
			if isStopped(s.stopped) {
				return errServerStopped
			}
			if isTempErr(err) {
				tempDelay := time.Millisecond * 5
				s.log.Tracef("accept err: %s; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return fmt.Errorf("accept err: %s", err)
		}
		if s.rwBufferSize > 0 {
			if err := conn.SetReadBuffer(s.rwBufferSize); err != nil {
				return fmt.Errorf("conn set read buffer err: %s", err)
			}
			if err := conn.SetWriteBuffer(s.rwBufferSize); err != nil {
				return fmt.Errorf("conn set write buffer err: %s", err)
			}
		}
		go s.handleConn(conn)
	}
}

// handleConn creates a new session according to conn,
// handles the message through the session in different goroutines,
// and waits until the session's closed.
func (s *TCPServer) handleConn(conn *net.TCPConn) {
	sess := session.NewTCPSession(conn, s.msgPacker, s.msgCodec)
	session.Sessions().Add(sess)
	go s.router.RouteLoop(sess)
	go sess.ReadLoop(s.readTimeout)
	go sess.WriteLoop(s.writeTimeout)
	sess.WaitUntilClosed()
	session.Sessions().Remove(sess.ID()) // session has been closed, remove it
	s.log.WithField("sid", sess.ID()).Tracef("session closed")
	if err := conn.Close(); err != nil {
		s.log.Tracef("connection close err: %s", err)
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
	s.log.Tracef("%d session(s) closed", closedNum)
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
