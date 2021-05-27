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

type Session struct {
	Id        string
	CreatedAt time.Time
	ClosedAt  time.Time
	Conn      net.Conn

	log *logrus.Entry

	connCloseOnce sync.Once
	connClosed    chan struct{}

	reqQueue chan *packet.Request
	ackQueue chan []byte

	MsgPacker packet.Packer // 解包和封包
	MsgCodec  packet.Codec  // encode/decode 包里的data
}

func New(conn net.Conn, packer packet.Packer, codec packet.Codec) *Session {
	return &Session{
		Id:         uuid.NewString(),
		CreatedAt:  time.Now(),
		Conn:       conn,
		connClosed: make(chan struct{}),
		log:        logger.Default.WithField("scope", "session"),
		reqQueue:   make(chan *packet.Request, 1024),
		ackQueue:   make(chan []byte, 1024),
		MsgPacker:  packer,
		MsgCodec:   codec,
	}
}

func (s *Session) WaitToClose() error {
	<-s.connClosed
	return s.Conn.Close()
}

func (s *Session) Close() {
	s.connCloseOnce.Do(func() {
		close(s.connClosed)
		close(s.reqQueue)
		close(s.ackQueue)
		s.ClosedAt = time.Now()
	})
}

func (s *Session) isClosed() bool {
	select {
	case <-s.connClosed:
		return true
	default:
		return false
	}
}

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
		data, err := s.MsgCodec.Decode(msg.GetData())
		if err != nil {
			s.log.Errorf("decode msg data err: %s", err)
			return
		}
		req := &packet.Request{
			Id:   msg.GetId(),
			Data: data,
		}
		s.reqQueue <- req
	}
}

func (s *Session) RecvReq() *packet.Request {
	if s.isClosed() {
		return nil
	}
	return <-s.reqQueue
}

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
	s.ackQueue <- msg
	return nil
}

func (s *Session) WriteLoop() {
	defer func() {
		s.Close()
		s.log.Warnf("write loop finished")
	}()
	for {
		if s.isClosed() {
			return
		}
		b := <-s.ackQueue
		if _, err := s.Conn.Write(b); err != nil {
			s.log.Errorf("conn write err: %s", err)
			return
		}
	}
}
