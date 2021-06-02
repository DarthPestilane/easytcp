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
	return &TcpSession{
		id:        uuid.NewString(),
		conn:      conn,
		closed:    make(chan struct{}),
		log:       logger.Default.WithField("scope", "session.TcpSession"),
		reqQueue:  make(chan *packet.Request, 1024),
		ackQueue:  make(chan []byte, 1024),
		msgPacker: packer,
		msgCodec:  codec,
	}
}

func (s *TcpSession) ID() string {
	return s.id
}

func (s *TcpSession) MsgPacker() packet.Packer {
	return s.msgPacker
}

func (s *TcpSession) MsgCodec() packet.Codec {
	return s.msgCodec
}

func (s *TcpSession) WaitUntilClosed() {
	<-s.closed
}

func (s *TcpSession) Close() error {
	var err error
	s.closeOnce.Do(func() {
		close(s.closed)
		close(s.reqQueue)
		close(s.ackQueue)
		if s.conn != nil {
			err = s.conn.Close()
		}
	})
	return err
}

func (s *TcpSession) isClosed() bool {
	select {
	case <-s.closed:
		return true
	default:
		return false
	}
}

// ReadLoop 阻塞式读消息，读到消息后，
// 通过 msgPacker 和 msgCodec 对原始消息进行处理
// 发送到对应的 channel 中，等待消费
func (s *TcpSession) ReadLoop() {
	defer func() {
		if err := s.Close(); err != nil {
			s.log.Tracef("conn close err: %s", err)
		}
	}()
	for {
		if s.isClosed() {
			return
		}
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

// RecvReq 接收请求
func (s *TcpSession) RecvReq() <-chan *packet.Request {
	return s.reqQueue
}

// SendResp 发送响应，
// resp 会经过 msgCodec 和 msgPacker 处理得到待写入的消息
func (s *TcpSession) SendResp(resp *packet.Response) error {
	if s.isClosed() {
		return fmt.Errorf("session closed")
	}
	if resp == nil {
		return fmt.Errorf("nil response")
	}
	data, err := s.msgCodec.Encode(resp.Data)
	if err != nil {
		return fmt.Errorf("encode response data err: %s", err)
	}
	msg, err := s.msgPacker.Pack(resp.Id, data)
	if err != nil {
		return fmt.Errorf("pack response data err: %s", err)
	}
	s.safelyPushAckQueue(msg)
	return nil
}

// WriteLoop 消费 ackQueue, 并写入连接
func (s *TcpSession) WriteLoop() {
	defer func() {
		if err := s.Close(); err != nil {
			s.log.Tracef("conn close err: %s", err)
		}
	}()
	for {
		if s.isClosed() {
			return
		}
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

func (s *TcpSession) safelyPushAckQueue(msg []byte) {
	defer func() {
		if r := recover(); r != nil {
			s.log.Tracef("push ackQueue panics: %+v", r)
		}
	}()
	s.ackQueue <- msg
}
