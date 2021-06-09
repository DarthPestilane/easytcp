package server

import (
	"fmt"
	"github.com/DarthPestilane/easytcp/logger"
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/DarthPestilane/easytcp/router"
	"github.com/DarthPestilane/easytcp/session"
	"github.com/sirupsen/logrus"
	"net"
)

// UDPServer is a server for UDP connections.
// UDPServer implements the Server interface.
type UDPServer struct {
	conn          *net.UDPConn
	rwBufferSize  int
	maxBufferSize int
	log           *logrus.Entry
	msgPacker     packet.Packer
	msgCodec      packet.Codec
	accepting     chan struct{}
	stopped       chan struct{}
	router        *router.Router
}

var _ Server = &UDPServer{}

// UDPOption is the option for UDPServer.
type UDPOption struct {
	MaxBufferSize int
	RWBufferSize  int
	MsgPacker     packet.Packer
	MsgCodec      packet.Codec
}

// NewUDPServer creates a UDPServer pointer according to opt.
func NewUDPServer(opt UDPOption) *UDPServer {
	if opt.MaxBufferSize <= 0 {
		opt.MaxBufferSize = 1024
	}
	if opt.MsgPacker == nil {
		opt.MsgPacker = &packet.DefaultPacker{}
	}
	if opt.MsgCodec == nil {
		opt.MsgCodec = &packet.StringCodec{}
	}
	return &UDPServer{
		log:           logger.Default.WithField("scope", "server.UDPServer"),
		rwBufferSize:  opt.RWBufferSize,
		msgPacker:     opt.MsgPacker,
		msgCodec:      opt.MsgCodec,
		maxBufferSize: opt.MaxBufferSize,
		accepting:     make(chan struct{}),
		stopped:       make(chan struct{}),
		router:        router.New(),
	}
}

// Serve implements the Server Serve method.
// Serve starts to listen UDP, and keep reading from UDP connection in a loop.
// The loop will break when error occurred and the error will be returned.
func (s *UDPServer) Serve(addr string) error {
	address, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP("udp", address)
	if err != nil {
		return err
	}
	if s.rwBufferSize > 0 {
		if err := conn.SetReadBuffer(s.rwBufferSize); err != nil {
			return fmt.Errorf("conn set read buffer err: %s", err)
		}
		if err := conn.SetWriteBuffer(s.rwBufferSize); err != nil {
			return fmt.Errorf("conn set write buffer err: %s", err)
		}
	}
	s.conn = conn
	return s.acceptLoop()
}

// acceptLoop keeps reading bytes from UDP connection and handle bytes in goroutine.
// Returns error when error occurred.
func (s *UDPServer) acceptLoop() error {
	close(s.accepting)
	buff := make([]byte, s.maxBufferSize)
	for {
		n, remoteAddr, err := s.conn.ReadFromUDP(buff)
		if err != nil {
			return fmt.Errorf("read conn err: %s", err)
		}
		go s.handleIncomingMsg(buff[:n], remoteAddr)
	}
}

// handleIncomingMsg creates a session.UDPSession to handle the incoming msg.
// And starts routing the message to the handler.
// Session will close after finishing writing, or the server's closed.
func (s *UDPServer) handleIncomingMsg(msg []byte, addr *net.UDPAddr) {
	sess := session.NewUDP(s.conn, addr, s.msgPacker, s.msgCodec)
	defer func() { s.log.WithField("sid", sess.ID()).Tracef("session closed") }()

	go s.router.Loop(sess)
	if err := sess.ReadIncomingMsg(msg); err != nil {
		return
	}
	sess.Write(s.stopped)
	sess.Close()
}

// Stop implements the Server Stop method.
// Stop stops server by close the connection.
func (s *UDPServer) Stop() error {
	close(s.stopped)
	return s.conn.Close()
}

// AddRoute implements the Server AddRoute method.
// AddRoute registers message handler and middlewares to the router.
func (s *UDPServer) AddRoute(msgID uint, handler router.HandlerFunc, middlewares ...router.MiddlewareFunc) {
	s.router.Register(msgID, handler, middlewares...)
}

// Use implements the Server Use method.
// Use registers global middlewares to the router.
func (s *UDPServer) Use(middlewares ...router.MiddlewareFunc) {
	s.router.RegisterMiddleware(middlewares...)
}
