package easytcp

import (
	"fmt"
	"github.com/DarthPestilane/easytcp/message"
	"github.com/google/uuid"
	"net"
	"sync"
	"time"
)

// Session represents a TCP session.
type Session struct {
	id        string              // session's ID. it's a UUID
	conn      net.Conn            // tcp connection
	closed    chan struct{}       // to close()
	respQueue chan *message.Entry // response queue channel, pushed in SendResp() and popped in writeOutbound()
	packer    Packer              // to pack and unpack message
	codec     Codec               // encode/decode message data
	ctxPool   sync.Pool
}

// SessionOption is the extra options for Session.
type SessionOption struct {
	Packer        Packer
	Codec         Codec
	respQueueSize int
}

// newSession creates a new Session.
// Parameter conn is the TCP connection,
// opt includes packer, codec, and channel size.
// Returns a Session pointer.
func newSession(conn net.Conn, opt *SessionOption) *Session {
	id := uuid.NewString()
	return &Session{
		id:        id,
		conn:      conn,
		closed:    make(chan struct{}),
		respQueue: make(chan *message.Entry, opt.respQueueSize),
		packer:    opt.Packer,
		codec:     opt.Codec,
		ctxPool:   sync.Pool{New: func() interface{} { return new(Context) }},
	}
}

// ID returns the session's ID.
func (s *Session) ID() string {
	return s.id
}

// SendResp pushes response message entry to respQueue.
// Returns error if session is closed.
func (s *Session) SendResp(respMsg *message.Entry) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("sessions is closed")
		}
	}()

	select {
	case s.respQueue <- respMsg:
	case <-s.closed:
		close(s.respQueue)
		err = fmt.Errorf("sessions is closed")
	}

	return
}

// close closes the session.
func (s *Session) close() {
	defer func() { _ = recover() }()
	close(s.closed)
}

// readInbound reads message packet from connection in a loop.
// And send unpacked message to reqQueue, which will be consumed in router.
// The loop breaks if errors occurred or the session is closed.
func (s *Session) readInbound(reqQueue chan<- *Context, timeout time.Duration) {
	for {
		if timeout > 0 {
			if err := s.conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
				Log.Errorf("session %s set read deadline err: %s", s.id, err)
				break
			}
		}
		entry, err := s.packer.Unpack(s.conn)
		if err != nil {
			Log.Errorf("session %s unpack inbound packet err: %s", s.id, err)
			break
		}
		if entry == nil {
			continue
		}

		ctx := s.ctxPool.Get().(*Context)
		ctx.session = s
		ctx.reqMsgEntry = entry
		ctx.storage = nil // reset storage
		select {
		case reqQueue <- ctx:
		case <-s.closed:
			Log.Tracef("session %s readInbound exit because session is closed", s.id)
			return
		}
	}
	Log.Tracef("session %s readInbound exit because of error", s.id)
	s.close()
}

// writeOutbound fetches message from respQueue channel and writes to TCP connection in a loop.
// Parameter writeTimeout specified the connection writing timeout.
// The loop breaks if errors occurred, or the session is closed.
func (s *Session) writeOutbound(writeTimeout time.Duration) {
FOR:
	for {
		select {
		case <-s.closed:
			Log.Tracef("session %s writeOutbound exit because session is closed", s.id)
			return
		case respMsg, ok := <-s.respQueue:
			if !ok {
				Log.Tracef("session %s writeOutbound exit because session is closed", s.id)
				return
			}
			// pack message
			outboundMsg, err := s.packer.Pack(respMsg)
			if err != nil {
				Log.Errorf("session %s pack outbound message err: %s", s.id, err)
				continue
			}
			if outboundMsg == nil {
				continue
			}
			if writeTimeout > 0 {
				if err := s.conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
					Log.Errorf("session %s set write deadline err: %s", s.id, err)
					break FOR
				}
			}
			if _, err := s.conn.Write(outboundMsg); err != nil {
				Log.Errorf("session %s conn write err: %s", s.id, err)
				break FOR
			}
		}
	}
	s.close()
	Log.Tracef("session %s writeOutbound exit because of error", s.id)
}
