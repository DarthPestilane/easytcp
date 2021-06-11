package session

import (
	"fmt"
	"github.com/DarthPestilane/easytcp/logger"
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"net"
	"sync"
	"time"
)

// TCPSession represents a TCP session.
// Implements Session interface.
type TCPSession struct {
	id        string              // session's ID. it's a uuid
	conn      net.Conn            // tcp connection
	log       *logrus.Entry       // logger
	closeOnce sync.Once           // to make sure we can only close each session one time
	closed    chan struct{}       // to close()
	reqQueue  chan packet.Message // request queue channel, pushed in ReadLoop() and popped in router.Router
	ackQueue  chan []byte         // ack queue channel, pushed in SendResp() and popped in WriteLoop()
	msgPacker packet.Packer       // to pack and unpack message
	msgCodec  packet.Codec        // encode/decode message data
}

var _ Session = &TCPSession{}

// NewTCP creates a new TCPSession.
// Parameter conn is the TCP connection,
// packer and codec will be used to pack/unpack and encode/decode message.
// Returns a TCPSession pointer.
func NewTCP(conn net.Conn, packer packet.Packer, codec packet.Codec) *TCPSession {
	id := uuid.NewString()
	return &TCPSession{
		id:        id,
		conn:      conn,
		closed:    make(chan struct{}),
		log:       logger.Default.WithField("sid", id).WithField("scope", "session.TCPSession"),
		reqQueue:  make(chan packet.Message, 1024),
		ackQueue:  make(chan []byte, 1024),
		msgPacker: packer,
		msgCodec:  codec,
	}
}

// ID implements the Session ID method.
// Returns session's ID.
func (s *TCPSession) ID() string {
	return s.id
}

// MsgCodec implements the Session MsgCodec method.
// Returns the message codec bound to session.
func (s *TCPSession) MsgCodec() packet.Codec {
	return s.msgCodec
}

// RecvReq implements the Session RecvReq method.
// Returns reqQueue channel which contains packet.Message.
func (s *TCPSession) RecvReq() <-chan packet.Message {
	return s.reqQueue
}

// SendResp implements the Session SendResp method.
// Pack respMsg and push to ackQueue channel.
// It won't panic even when ackQueue channel is closed.
// It returns error when encode or pack failed.
func (s *TCPSession) SendResp(respMsg packet.Message) (closed bool, _ error) {
	ackMsg, err := s.msgPacker.Pack(respMsg)
	if err != nil {
		return false, fmt.Errorf("pack response data err: %s", err)
	}
	return !s.safelyPushAckQueue(ackMsg), nil
}

// Close closes the session by closing all the channels.
func (s *TCPSession) Close() {
	s.closeOnce.Do(func() {
		close(s.closed)
		close(s.reqQueue)
		close(s.ackQueue)
	})
}

// ReadLoop reads TCP connection, unpacks message packet
// to a packet.Message, and push to reqQueue channel.
// The above operations are in a loop.
// Parameter readTimeout specified the connection reading timeout.
// The loop will break if any error occurred, or the session is closed.
// After loop ended, this session will be closed.
func (s *TCPSession) ReadLoop(readTimeout time.Duration) {
	for {
		if readTimeout > 0 {
			if err := s.conn.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
				s.log.Tracef("set read deadline err: %s", err)
				break
			}
		}
		msg, err := s.msgPacker.Unpack(s.conn)
		if err != nil {
			s.log.Tracef("unpack incoming message err: %s", err)
			break
		}
		if !s.safelyPushReqQueue(msg) {
			break
		}
	}
	s.log.Tracef("read loop exit")
	s.Close()
}

// WriteLoop fetches message from ackQueue channel and writes to TCP connection.
// The above operations are in a loop.
// Parameter writeTimeout specified the connection writing timeout.
// The loop will break if any error occurred, or the session is closed.
// After loop ended, this session will be closed.
func (s *TCPSession) WriteLoop(writeTimeout time.Duration) {
	for {
		msg, ok := <-s.ackQueue
		if !ok {
			break
		}
		if writeTimeout > 0 {
			if err := s.conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
				s.log.Tracef("set write deadline err: %s", err)
				break
			}
		}
		if _, err := s.conn.Write(msg); err != nil {
			s.log.Tracef("conn write err: %s", err)
			break
		}
	}
	s.log.Tracef("write loop exit")
	s.Close()
}

// WaitUntilClosed waits until the session is closed.
func (s *TCPSession) WaitUntilClosed() {
	<-s.closed
}

func (s *TCPSession) safelyPushReqQueue(reqMsg packet.Message) (ok bool) {
	ok = true
	defer func() {
		if r := recover(); r != nil {
			ok = false
			s.log.Tracef("push reqQueue panics: %+v", r)
		}
	}()
	s.reqQueue <- reqMsg
	return ok
}

func (s *TCPSession) safelyPushAckQueue(ackMsg []byte) (ok bool) {
	ok = true
	defer func() {
		if r := recover(); r != nil {
			ok = false
			s.log.Tracef("push ackQueue panics: %+v", r)
		}
	}()
	s.ackQueue <- ackMsg
	return ok
}
