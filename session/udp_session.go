package session

import (
	"bytes"
	"fmt"
	"github.com/DarthPestilane/easytcp/logger"
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"net"
	"sync"
)

type UdpSession struct {
	id         string       // 会话的ID，uuid形式
	conn       *net.UDPConn // 网络连接
	log        *logrus.Entry
	closeOnce  sync.Once
	closed     chan struct{} // to close()
	reqQueue   chan *packet.Request
	ackQueue   chan []byte
	msgPacker  packet.Packer // 拆包和封包
	msgCodec   packet.Codec  // encode/decode 包里的data
	remoteAddr *net.UDPAddr
}

func NewUdp(conn *net.UDPConn, addr *net.UDPAddr, packer packet.Packer, codec packet.Codec) *UdpSession {
	id := uuid.NewString()
	return &UdpSession{
		id:         id,
		conn:       conn,
		closed:     make(chan struct{}),
		log:        logger.Default.WithField("sid", id).WithField("scope", "session.UdpSession"),
		reqQueue:   make(chan *packet.Request),
		ackQueue:   make(chan []byte),
		msgPacker:  packer,
		msgCodec:   codec,
		remoteAddr: addr,
	}
}

func (s *UdpSession) ID() string {
	return s.id
}

func (s *UdpSession) MsgCodec() packet.Codec {
	return s.msgCodec
}

func (s *UdpSession) RecvReq() <-chan *packet.Request {
	return s.reqQueue
}

func (s *UdpSession) SendResp(resp *packet.Response) (closed bool, _ error) {
	if s.isClosed() {
		return true, nil
	}
	data, err := s.msgCodec.Encode(resp.Data)
	if err != nil {
		return false, fmt.Errorf("encode response data err: %s", err)
	}
	msg, err := s.msgPacker.Pack(resp.Id, data)
	if err != nil {
		return false, fmt.Errorf("pack response data err: %s", err)
	}
	return !s.safelyPushAckQueue(msg), nil
}

func (s *UdpSession) ReadIncomingMsg(inMsg []byte) {
	msg, err := s.msgPacker.Unpack(bytes.NewReader(inMsg))
	if err != nil {
		s.log.Tracef("unpack incoming message err: %s", err)
		return
	}
	req := &packet.Request{
		Id:      msg.GetId(),
		RawSize: msg.GetSize(),
		RawData: msg.GetData(),
	}
	s.safelyPushReqQueue(req)
}

func (s *UdpSession) Write() {
	msg, ok := <-s.ackQueue
	if !ok {
		return
	}
	if _, err := s.conn.WriteToUDP(msg, s.remoteAddr); err != nil {
		s.log.Tracef("conn write err: %s", err)
		return
	}
}

func (s *UdpSession) safelyPushReqQueue(req *packet.Request) {
	defer func() {
		if r := recover(); r != nil {
			s.log.Tracef("push reqQueue panics: %+v", r)
		}
	}()
	s.reqQueue <- req
}

func (s *UdpSession) safelyPushAckQueue(msg []byte) (ok bool) {
	ok = true
	defer func() {
		if r := recover(); r != nil {
			ok = false
			s.log.Tracef("push ackQueue panics: %+v", r)
		}
	}()
	s.ackQueue <- msg
	return ok
}

func (s *UdpSession) Close() {
	s.closeOnce.Do(func() {
		close(s.closed)
		close(s.reqQueue)
		close(s.ackQueue)
	})
}

func (s *UdpSession) isClosed() bool {
	select {
	case <-s.closed:
		return true
	default:
		return false
	}
}
