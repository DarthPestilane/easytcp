package easytcp

import (
	"github.com/google/uuid"
	"net"
	"sync"
	"time"
)

// Session represents a TCP session.
type Session interface {
	// ID returns current session's id.
	ID() string

	// Send sends the ctx to the respQueue.
	Send(ctx Context) bool

	// Codec returns the codec, can be nil.
	Codec() Codec

	// Close closes session.
	Close()

	// NewContext creates a Context.
	NewContext() Context
}

type session struct {
	id        string        // session's ID. it's a UUID
	conn      net.Conn      // tcp connection
	closed    chan struct{} // to close()
	closeOne  sync.Once     // ensure one session only close once
	respQueue chan Context  // response queue channel, pushed in Send() and popped in writeOutbound()
	packer    Packer        // to pack and unpack message
	codec     Codec         // encode/decode message data
	ctxPool   sync.Pool     // router context pool
}

// sessionOption is the extra options for session.
type sessionOption struct {
	Packer        Packer
	Codec         Codec
	respQueueSize int
}

// newSession creates a new session.
// Parameter conn is the TCP connection,
// opt includes packer, codec, and channel size.
// Returns a session pointer.
func newSession(conn net.Conn, opt *sessionOption) *session {
	return &session{
		id:        uuid.NewString(),
		conn:      conn,
		closed:    make(chan struct{}),
		respQueue: make(chan Context, opt.respQueueSize),
		packer:    opt.Packer,
		codec:     opt.Codec,
		ctxPool:   sync.Pool{New: func() interface{} { return NewContext() }},
	}
}

// ID returns the session's ID.
func (s *session) ID() string {
	return s.id
}

// Send pushes response message entry to respQueue.
// Returns error if session is closed.
func (s *session) Send(ctx Context) (ok bool) {
	select {
	case <-s.closed:
		return false
	case s.respQueue <- ctx:
		return true
	}
}

// Codec implements Session Codec.
func (s *session) Codec() Codec {
	return s.codec
}

// Close closes the session, but doesn't close the connection.
// The connection will be closed in the server once the session's closed.
func (s *session) Close() {
	s.closeOne.Do(func() { close(s.closed) })
}

// NewContext creates a Context from pool.
func (s *session) NewContext() Context {
	return s.ctxPool.Get().(*routeContext)
}

// readInbound reads message packet from connection in a loop.
// And send unpacked message to reqQueue, which will be consumed in router.
// The loop breaks if errors occurred or the session is closed.
func (s *session) readInbound(router *Router, timeout time.Duration) {
	for {
		select {
		case <-s.closed:
			return
		default:
		}
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

		// don't block the loop.
		go func() {
			ctx := s.NewContext().(*routeContext)
			ctx.reset(s, reqEntry)
			router.handleRequest(ctx)
			s.Send(ctx)
		}()
	}
	Log.Tracef("session %s readInbound exit because of error", s.id)
	s.Close()
}

// writeOutbound fetches message from respQueue channel and writes to TCP connection in a loop.
// Parameter writeTimeout specified the connection writing timeout.
// The loop breaks if errors occurred, or the session is closed.
func (s *session) writeOutbound(writeTimeout time.Duration, attemptTimes int) {
	for {
		var ctx Context
		select {
		case <-s.closed:
			return
		case ctx = <-s.respQueue:
		}

		outboundMsg, err := s.packResponse(ctx)
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
				break
			}
		}

		if err := s.attemptConnWrite(outboundMsg, attemptTimes); err != nil {
			Log.Errorf("session %s conn write err: %s", s.id, err)
			break
		}
	}
	s.Close()
	Log.Tracef("session %s writeOutbound exit because of error", s.id)
}

func (s *session) attemptConnWrite(outboundMsg []byte, attemptTimes int) (err error) {
	for i := 0; i < attemptTimes; i++ {
		time.Sleep(tempErrDelay * time.Duration(i))
		_, err = s.conn.Write(outboundMsg)

		// breaks if err is not nil or it's the last attempt.
		if err == nil || i == attemptTimes-1 {
			break
		}

		// check if err is `net.Error`
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
		break // if err is not temporary, break the loop.
	}
	return
}

func (s *session) packResponse(ctx Context) ([]byte, error) {
	defer s.ctxPool.Put(ctx)
	if ctx.Response() == nil {
		return nil, nil
	}
	return s.packer.Pack(ctx.Response())
}
