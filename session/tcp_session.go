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
	data, err := s.msgCodec.Encode(resp.Data)
	if err != nil {
		return false, fmt.Errorf("encode response data err: %s", err)
	}
	msg, err := s.msgPacker.Pack(resp.Id, data)
	if err != nil {
		return false, fmt.Errorf("pack response data err: %s", err)
	}
	ok := s.safelyPushAckQueue(msg)
	if !ok {
		s.Close()
		return true, nil
	}
	return false, nil
}

func (s *TcpSession) Close() {
	s.closeOnce.Do(func() {
		close(s.closed)
		close(s.reqQueue)
		close(s.ackQueue)
	})
}

func (s *TcpSession) ReadLoop() {
	for {
		msg, err := s.msgPacker.Unpack(s.conn)
		if err != nil {
			s.log.Tracef("unpack incoming message err:%s", err)
			break
		}
		req := &packet.Request{
			Id:      msg.GetId(),
			RawSize: msg.GetSize(),
			RawData: msg.GetData(),
		}
		if !s.safelyPushReqQueue(req) {
			break
		}
	}
	s.log.Tracef("read loop exit")
	s.Close()
}

func (s *TcpSession) WriteLoop() {
	for {
		msg, ok := <-s.ackQueue
		if !ok {
			break
		}
		if _, err := s.conn.Write(msg); err != nil {
			s.log.Tracef("conn write err: %s", err)
			break
		}
	}
	s.log.Tracef("write loop exit")
	s.Close()
}

func (s *TcpSession) safelyPushReqQueue(req *packet.Request) (ok bool) {
	ok = true
	defer func() {
		if r := recover(); r != nil {
			ok = false
			s.log.Tracef("push reqQueue panics: %+v", r)
		}
	}()
	s.reqQueue <- req
	return ok
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
