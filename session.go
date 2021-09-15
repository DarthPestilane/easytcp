package easytcp

import (
	"fmt"
	"github.com/google/uuid"
	"net"
	"sync"
	"time"
)

// Session represents a TCP session.
type Session struct {
	id        string        // session's ID. it's a UUID
	conn      net.Conn      // tcp connection
	closed    chan struct{} // to close()
	respQueue chan *Context // response queue channel, pushed in SendResp() and popped in writeOutbound()
	packer    Packer        // to pack and unpack message
	codec     Codec         // encode/decode message data
	ctxPool   sync.Pool     // router context pool
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
	return &Session{
		id:        uuid.NewString(),
		conn:      conn,
		closed:    make(chan struct{}),
		respQueue: make(chan *Context, opt.respQueueSize),
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
func (s *Session) SendResp(ctx *Context) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("sessions is closed")
		}
	}()

	select {
	case s.respQueue <- ctx:
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
		reqEntry, err := s.packer.Unpack(s.conn)
		if err != nil {
			Log.Errorf("session %s unpack inbound packet err: %s", s.id, err)
			break
		}
		if reqEntry == nil {
			continue
		}

		ctx := s.ctxPool.Get().(*Context)
		ctx.reset(s, reqEntry)

		if !s.sendReq(ctx, reqQueue) {
			Log.Tracef("session %s readInbound exit because session is closed", s.id)
			return
		}
	}
	Log.Tracef("session %s readInbound exit because of error", s.id)
	s.close()
}

func (s *Session) sendReq(ctx *Context, reqQueue chan<- *Context) (ok bool) {
	defer func() {
		if r := recover(); r != nil {
			ok = false
		}
	}()
	select {
	case reqQueue <- ctx:
		ok = true
	case <-s.closed:
		Log.Tracef("session %s readInbound exit because session is closed", s.id)
		ok = false
	}
	return
}

// writeOutbound fetches message from respQueue channel and writes to TCP connection in a loop.
// Parameter writeTimeout specified the connection writing timeout.
// The loop breaks if errors occurred, or the session is closed.
func (s *Session) writeOutbound(writeTimeout time.Duration) {
LOOP:
	for {
		select {
		case <-s.closed:
			Log.Tracef("session %s writeOutbound exit because session is closed", s.id)
			return
		case ctx, ok := <-s.respQueue:
			if !ok {
				Log.Tracef("session %s writeOutbound exit because session is closed", s.id)
				return
			}

			outboundMsg, err := s.pack(ctx)
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
					break LOOP
				}
			}

			if err := s.tryConnWrite(outboundMsg, 10); err != nil {
				Log.Errorf("session %s conn write err: %s", s.id, err)
				break LOOP
			}
		}
	}
	s.close()
	Log.Tracef("session %s writeOutbound exit because of error", s.id)
}

func (s *Session) tryConnWrite(outboundMsg []byte, maxTries int) (err error) {
	if maxTries <= 0 {
		maxTries = 1
	}
	for i := 0; i < maxTries; i++ {
		time.Sleep(tempErrDelay * time.Duration(i))
		_, err = s.conn.Write(outboundMsg)

		if err == nil {
			break
		}
		if i == maxTries-1 { // if it's the last loop
			break
		}

		// check net.Error
		ne, ok := err.(net.Error)
		if !ok {
			break
		}
		if ne.Timeout() {
			break
		}
		if ne.Temporary() {
			Log.Errorf("session %s conn write err: %s; retrying in %s", s.id, err, tempErrDelay*time.Duration(i+1))
			continue
		}
		break
	}
	return
}

func (s *Session) pack(ctx *Context) ([]byte, error) {
	defer s.ctxPool.Put(ctx)
	if ctx.respEntry == nil {
		return nil, nil
	}
	return s.packer.Pack(ctx.respEntry)
}
