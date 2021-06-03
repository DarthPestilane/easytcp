package session

import (
	"fmt"
	"github.com/DarthPestilane/easytcp/logger"
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"net"
	"sync"
)

// TcpSession 会话，负责读写和关闭连接
type TcpSession struct {
	id        string   // 会话的ID，uuid形式
	conn      net.Conn // 网络连接
	log       *logrus.Entry
	closeOnce sync.Once
	closed    chan struct{} // to close()
	reqQueue  chan *packet.Request
	ackQueue  chan []byte
	msgPacker packet.Packer // 拆包和封包
	msgCodec  packet.Codec  // encode/decode 包里的data
}

// NewTcp 创建一个会话
func NewTcp(conn net.Conn, packer packet.Packer, codec packet.Codec) *TcpSession {
	id := uuid.NewString()
	return &TcpSession{
		id:        id,
		conn:      conn,
		closed:    make(chan struct{}),
		log:       logger.Default.WithField("sid", id).WithField("scope", "session.TcpSession"),
		reqQueue:  make(chan *packet.Request, 1024),
		ackQueue:  make(chan []byte, 1024),
		msgPacker: packer,
		msgCodec:  codec,
	}
}

func (s *TcpSession) ID() string {
	return s.id
}

func (s *TcpSession) MsgCodec() packet.Codec {
	return s.msgCodec
}

func (s *TcpSession) RecvReq() <-chan *packet.Request {
	return s.reqQueue
}

func (s *TcpSession) SendResp(resp *packet.Response) (closed bool, _ error) {
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

func (s *TcpSession) Close() {
	s.closeOnce.Do(func() {
		close(s.closed)
		close(s.reqQueue)
		close(s.ackQueue)
	})
}

func (s *TcpSession) ReadLoop() {
	defer s.Close()
	for {
		msg, err := s.msgPacker.Unpack(s.conn)
		if err != nil {
			s.log.Tracef("unpack incoming message err:%s", err)
			return
		}
		req := &packet.Request{
			Id:      msg.GetId(),
			RawSize: msg.GetSize(),
			RawData: msg.GetData(),
		}
		s.safelyPushReqQueue(req)
	}
}

func (s *TcpSession) WriteLoop() {
	defer s.Close()
	for {
		msg, ok := <-s.ackQueue
		if !ok {
			return
		}
		if _, err := s.conn.Write(msg); err != nil {
			s.log.Tracef("conn write err: %s", err)
			return
		}
	}
}

func (s *TcpSession) safelyPushReqQueue(req *packet.Request) {
	defer func() {
		if r := recover(); r != nil {
			s.log.Tracef("push reqQueue panics: %+v", r)
		}
	}()
	s.reqQueue <- req
}

func (s *TcpSession) safelyPushAckQueue(msg []byte) (ok bool) {
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

func (s *TcpSession) WaitUntilClosed() {
	<-s.closed
}

func (s *TcpSession) isClosed() bool {
	select {
	case <-s.closed:
		return true
	default:
		return false
	}
}
