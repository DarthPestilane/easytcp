package session

import (
	"fmt"
	"github.com/DarthPestilane/easytcp/logger"
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/google/uuid"
	"net"
	"sync"
	"time"
)

// TCPSession represents a TCP session.
// Implements Session interface.
type TCPSession struct {
	id        string                    // session's ID. it's a uuid
	conn      net.Conn                  // tcp connection
	closeOnce sync.Once                 // to make sure we can only close each session one time
	closed    chan struct{}             // to close()
	reqQueue  chan *packet.MessageEntry // request queue channel, pushed in ReadLoop() and popped in router.Router
	respQueue chan *packet.MessageEntry // response queue channel, pushed in SendResp() and popped in WriteLoop()
	packer    packet.Packer             // to pack and unpack message
	codec     packet.Codec              // encode/decode message data
}

var _ Session = &TCPSession{}

// TCPSessionOption is the extra options for TCPSession.
type TCPSessionOption struct {
	Packer          packet.Packer
	Codec           packet.Codec
	ReadBufferSize  int
	WriteBufferSize int
}

// NewTCPSession creates a new TCPSession.
// Parameter conn is the TCP connection,
// opt includes packer, codec, and channel size.
// Returns a TCPSession pointer.
func NewTCPSession(conn net.Conn, opt *TCPSessionOption) *TCPSession {
	id := uuid.NewString()
	return &TCPSession{
		id:        id,
		conn:      conn,
		closed:    make(chan struct{}),
		reqQueue:  make(chan *packet.MessageEntry, opt.ReadBufferSize),
		respQueue: make(chan *packet.MessageEntry, opt.WriteBufferSize),
		packer:    opt.Packer,
		codec:     opt.Codec,
	}
}

// ID implements the Session ID method.
// Returns session's ID.
func (s *TCPSession) ID() string {
	return s.id
}

// Codec implements the Session Codec method.
// Returns the message codec bound to session.
func (s *TCPSession) Codec() packet.Codec {
	return s.codec
}

// RecvReq implements the Session RecvReq method.
// Returns reqQueue channel which contains packet.Message.
func (s *TCPSession) RecvReq() <-chan *packet.MessageEntry {
	return s.reqQueue
}

// SendResp implements the Session SendResp method.
// If respQueue is closed, returns false.
func (s *TCPSession) SendResp(respMsg *packet.MessageEntry) error {
	if !s.safelyPushRespQueue(respMsg) {
		return fmt.Errorf("session's closed")
	}
	return nil
}

// Close closes the session by closing all the channels.
func (s *TCPSession) Close() {
	s.closeOnce.Do(func() {
		close(s.closed)
		close(s.reqQueue)
		close(s.respQueue)
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
				logger.Log.Tracef("set read deadline err: %s", err)
				break
			}
		}
		msg, err := s.packer.Unpack(s.conn)
		if err != nil {
			logger.Log.Tracef("unpack incoming message err: %s", err)
			break
		}
		if !s.safelyPushReqQueue(msg) {
			break
		}
	}
	logger.Log.Tracef("read loop exit")
	s.Close()
}

// WriteLoop fetches message from respQueue channel and writes to TCP connection.
// The above operations are in a loop.
// Parameter writeTimeout specified the connection writing timeout.
// The loop will break if any error occurred, or the session is closed.
// After loop ended, this session will be closed.
func (s *TCPSession) WriteLoop(writeTimeout time.Duration) {
	for {
		respMsg, ok := <-s.respQueue
		if !ok {
			break
		}
		// pack message
		ackMsg, err := s.packer.Pack(respMsg)
		if err != nil {
			logger.Log.Tracef("pack response message err: %s", err)
			continue
		}
		if writeTimeout > 0 {
			if err := s.conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
				logger.Log.Tracef("set write deadline err: %s", err)
				break
			}
		}
		if _, err := s.conn.Write(ackMsg); err != nil {
			logger.Log.Tracef("conn write err: %s", err)
			break
		}
	}
	logger.Log.Tracef("write loop exit")
	s.Close()
}

// WaitUntilClosed waits until the session is closed.
func (s *TCPSession) WaitUntilClosed() {
	<-s.closed
}

func (s *TCPSession) safelyPushReqQueue(reqMsg *packet.MessageEntry) (ok bool) {
	ok = true
	defer func() {
		if r := recover(); r != nil {
			ok = false
			logger.Log.Tracef("push reqQueue panics: %+v", r)
		}
	}()
	s.reqQueue <- reqMsg
	return ok
}

func (s *TCPSession) safelyPushRespQueue(respMsg *packet.MessageEntry) (ok bool) {
	ok = true
	defer func() {
		if r := recover(); r != nil {
			ok = false
			logger.Log.Tracef("push respQueue panics: %+v", r)
		}
	}()
	s.respQueue <- respMsg
	return ok
}
