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

type UdpServer struct {
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

var _ Server = &UdpServer{}

type UdpOption struct {
	MaxBufferSize int
	RWBufferSize  int
	MsgPacker     packet.Packer
	MsgCodec      packet.Codec
}

func NewUdpServer(opt UdpOption) *UdpServer {
	if opt.MaxBufferSize <= 0 {
		opt.MaxBufferSize = 1024
	}
	if opt.MsgPacker == nil {
		opt.MsgPacker = &packet.DefaultPacker{}
	}
	if opt.MsgCodec == nil {
		opt.MsgCodec = &packet.StringCodec{}
	}
	return &UdpServer{
		log:           logger.Default.WithField("scope", "server.UdpServer"),
		rwBufferSize:  opt.RWBufferSize,
		msgPacker:     opt.MsgPacker,
		msgCodec:      opt.MsgCodec,
		maxBufferSize: opt.MaxBufferSize,
		accepting:     make(chan struct{}),
		stopped:       make(chan struct{}),
		router:        router.New(),
	}
}

func (s *UdpServer) Serve(addr string) error {
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

func (s *UdpServer) acceptLoop() error {
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

func (s *UdpServer) handleIncomingMsg(msg []byte, addr *net.UDPAddr) {
	sess := session.NewUdp(s.conn, addr, s.msgPacker, s.msgCodec)
	defer func() { s.log.WithField("sid", sess.ID()).Tracef("session closed") }()

	go s.router.Loop(sess)
	if err := sess.ReadIncomingMsg(msg); err != nil {
		return
	}
	sess.Write(s.stopped)
	sess.Close()
}

func (s *UdpServer) Stop() error {
	close(s.stopped)
	return s.conn.Close()
}

func (s *UdpServer) AddRoute(msgId uint, handler router.HandlerFunc, middlewares ...router.MiddlewareFunc) {
	s.router.Register(msgId, handler, middlewares...)
}

func (s *UdpServer) Use(middlewares ...router.MiddlewareFunc) {
	s.router.RegisterMiddleware(middlewares...)
}
