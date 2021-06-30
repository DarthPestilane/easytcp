package session

import (
	"bytes"
	"fmt"
	"github.com/DarthPestilane/easytcp/logger"
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"net"
)

// UDPSession represents a UDP session.
// Implements Session interface.
type UDPSession struct {
	id         string                    // session's id. a uuid
	conn       *net.UDPConn              // udp connection
	log        *logrus.Entry             // logger
	closed     chan struct{}             // represents whether the session is closed. will be closed in Close() method
	reqQueue   chan *packet.MessageEntry // a non-buffer channel, pushed in ReadIncomingMsg(), popped in router.Router
	respQueue  chan *packet.MessageEntry // a non-buffer channel, pushed in SendResp(), popped in Write()
	msgPacker  packet.Packer             // pack/unpack message packet
	msgCodec   packet.Codec              // encode/decode message data
	remoteAddr *net.UDPAddr              // UDP remote address, used to conn.WriteToUDP(remoteAddr)
}

var _ Session = &UDPSession{}

// NewUDPSession creates a new UDPSession.
// Parameter conn is the UDP connection, addr will be used as remote UDP peer address to write to,
// packer and codec will be used to pack/unpack and encode/decode message.
// Returns a UDPSession pointer.
func NewUDPSession(conn *net.UDPConn, addr *net.UDPAddr, packer packet.Packer, codec packet.Codec) *UDPSession {
	id := uuid.NewString()
	return &UDPSession{
		id:         id,
		conn:       conn,
		closed:     make(chan struct{}),
		log:        logger.Default.WithField("sid", id).WithField("scope", "session.UDPSession"),
		reqQueue:   make(chan *packet.MessageEntry),
		respQueue:  make(chan *packet.MessageEntry),
		msgPacker:  packer,
		msgCodec:   codec,
		remoteAddr: addr,
	}
}

// ID implements the Session ID method.
// Returns session's ID.
func (s *UDPSession) ID() string {
	return s.id
}

// MsgCodec implements the Session MsgCodec method.
// Returns the message codec bound to session.
func (s *UDPSession) MsgCodec() packet.Codec {
	return s.msgCodec
}

// RecvReq implements the Session RecvReq method.
// Returns reqQueue channel which contains packet.Message.
func (s *UDPSession) RecvReq() <-chan *packet.MessageEntry {
	return s.reqQueue
}

// SendResp implements the Session SendResp method.
// If respQueue channel is closed, returns false.
func (s *UDPSession) SendResp(respMsg *packet.MessageEntry) error {
	if !s.safelyPushRespQueue(respMsg) {
		return fmt.Errorf("session's closed")
	}
	return nil
}

// ReadIncomingMsg reads and unpacks the incoming message packet inMsg
// to a packet.Message and push to reqQueue.
// Returns error when unpack failed.
func (s *UDPSession) ReadIncomingMsg(inMsg []byte) error {
	reqMsg, err := s.msgPacker.Unpack(bytes.NewReader(inMsg))
	if err != nil {
		return fmt.Errorf("unpack incoming message err: %s", err)
	}
	s.safelyPushReqQueue(reqMsg)
	return nil
}

// Write writes the message to a UDP peer.
// Will stop as soon as <-done or respQueue closed,
// or when connection failed to write.
func (s *UDPSession) Write(done <-chan struct{}) {
	select {
	case <-done:
		return
	case respMsg, ok := <-s.respQueue:
		if !ok {
			return
		}
		ackMsg, err := s.msgPacker.Pack(respMsg)
		if err != nil {
			s.log.Tracef("pack response message err: %s", err)
			return
		}
		if _, err := s.conn.WriteToUDP(ackMsg, s.remoteAddr); err != nil {
			s.log.Tracef("conn write err: %s", err)
			return
		}
	}
}

// Close closes the session by closing all the channels.
// NOT safe in concurrency, each session should call Close() only for one time.
func (s *UDPSession) Close() {
	close(s.closed)
	close(s.reqQueue)
	close(s.respQueue)
}

func (s *UDPSession) safelyPushReqQueue(reqMsg *packet.MessageEntry) {
	defer func() {
		if r := recover(); r != nil {
			s.log.Tracef("push reqQueue panics: %+v", r)
		}
	}()
	s.reqQueue <- reqMsg
}

func (s *UDPSession) safelyPushRespQueue(respMsg *packet.MessageEntry) (ok bool) {
	ok = true
	defer func() {
		if r := recover(); r != nil {
			ok = false
			s.log.Tracef("push respQueue panics: %+v", r)
		}
	}()
	s.respQueue <- respMsg
	return ok
}
