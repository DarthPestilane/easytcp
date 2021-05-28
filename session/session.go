package session

import (
	"github.com/DarthPestilane/easytcp/logger"
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"net"
	"sync"
	"time"
)

// Session 会话，负责读写和关闭连接
type Session struct {
	Id        string    // 会话的ID，uuid形式
	CreatedAt time.Time // 创建时间
	ClosedAt  time.Time // 关闭时间
	Conn      net.Conn  // 网络连接

	log *logrus.Entry

	closeOnce sync.Once
	closed    chan struct{} // to close()

	reqQueue chan *packet.Request
	ackQueue chan []byte

	MsgPacker packet.Packer // 拆包和封包
	MsgCodec  packet.Codec  // encode/decode 包里的data
}

// New 创建一个会话
func New(conn net.Conn, packer packet.Packer, codec packet.Codec) *Session {
	return &Session{
		Id:        uuid.NewString(),
		CreatedAt: time.Now(),
		Conn:      conn,
		closed:    make(chan struct{}),
		log:       logger.Default.WithField("scope", "session.Session"),
		reqQueue:  make(chan *packet.Request, 1024),
		ackQueue:  make(chan []byte, 1024),
		MsgPacker: packer,
		MsgCodec:  codec,
	}
}

// WaitToClose 等待会话关闭，关闭底层连接
func (s *Session) WaitToClose() error {
	<-s.closed
	return s.Conn.Close()
}

// Close 关闭会话，通过 close(ch) 方式
func (s *Session) Close() {
	s.closeOnce.Do(func() {
		close(s.closed)

		// and close other channels
		close(s.reqQueue)
		close(s.ackQueue)

		s.ClosedAt = time.Now()
	})
}

func (s *Session) isClosed() bool {
	select {
	case <-s.closed:
		return true
	default:
		return false
	}
}

// ReadLoop 阻塞式读消息，读到消息后，
// 通过 MsgPacker 和 MsgCodec 对原始消息进行处理
// 发送到对应的 channel 中，等待消费
func (s *Session) ReadLoop() {
	defer func() {
		s.Close()
		s.log.Warnf("read loop finished")
	}()
	for {
		if s.isClosed() {
			return
		}
		msg, err := s.MsgPacker.Unpack(s.Conn)
		if err != nil {
			s.log.Errorf("unpack msg err:%s", err)
			return
		}
		decodedData, err := s.MsgCodec.Decode(msg.GetData())
		if err != nil {
			s.log.Errorf("decode msg data err: %s", err)
			return
		}
		req := &packet.Request{
			Id:      msg.GetId(),
			RawSize: msg.GetSize(),
			Data:    decodedData,
			RawData: msg.GetData(),
		}
		s.safelyPushReqQueue(req)
	}
}

// RecvReq 接收请求
func (s *Session) RecvReq() (*packet.Request, bool) {
	if s.isClosed() {
		return nil, false
	}
	req, ok := <-s.reqQueue
	return req, ok
}

// SendResp 发送响应，
// resp 会经过 MsgCodec 和 MsgPacker 处理得到待写入的消息
func (s *Session) SendResp(resp *packet.Response) error {
	if s.isClosed() {
		return nil
	}
	data, err := s.MsgCodec.Encode(resp.Data)
	if err != nil {
		return err
	}
	msg, err := s.MsgPacker.Pack(resp.Id, data)
	if err != nil {
		return err
	}
	s.safelyPushAckQueue(msg)
	return nil
}

// WriteLoop 消费 ackQueue, 并写入连接
func (s *Session) WriteLoop() {
	defer func() {
		s.Close()
		s.log.Warnf("write loop finished")
	}()
	for {
		if s.isClosed() {
			return
		}
		msg, ok := <-s.ackQueue
		if !ok {
			return
		}
		if _, err := s.Conn.Write(msg); err != nil {
			s.log.Errorf("conn write err: %s", err)
			return
		}
	}
}

func (s *Session) safelyPushReqQueue(req *packet.Request) {
	defer func() {
		if r := recover(); r != nil {
			s.log.Errorf("push reqQueue panics: %s", r)
		}
	}()
	s.reqQueue <- req
}

func (s *Session) safelyPushAckQueue(msg []byte) {
	defer func() {
		if r := recover(); r != nil {
			s.log.Errorf("push ackQueue panics: %s", r)
		}
	}()
	s.ackQueue <- msg
}
