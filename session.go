package easytcp

import (
	"fmt"
	"github.com/google/uuid"
	"io"
	"net"
	"sync"
	"time"
)

// Session represents a TCP session.
type Session interface {
	// ID returns current session's id.
	ID() interface{}

	// SetID sets current session's id.
	SetID(id interface{})

	// Send sends the ctx to the respStream.
	Send(ctx Context) bool

	// Codec returns the codec, can be nil.
	Codec() Codec

	// Close closes current session.
	Close()

	// AllocateContext gets a Context ships with current session.
	AllocateContext() Context

	// Conn returns the underlined connection.
	Conn() net.Conn

	// AfterCreateHook blocks until session's on-create hook triggered.
	AfterCreateHook() <-chan struct{}

	// AfterCloseHook blocks until session's on-close hook triggered.
	AfterCloseHook() <-chan struct{}
}

type session struct {
	id               interface{}   // session's ID.
	conn             net.Conn      // tcp connection
	closedC          chan struct{} // to close when read/write loop stopped
	closeOnce        sync.Once     // ensure one session only close once
	afterCreateHookC chan struct{} // to close after session's on-create hook triggered
	afterCloseHookC  chan struct{} // to close after session's on-close hook triggered
	respStream       chan Context  // response queue channel, pushed in Send() and popped in writeOutbound()
	packer           Packer        // to pack and unpack message
	codec            Codec         // encode/decode message data
	ctxPool          sync.Pool     // router context pool
	asyncRouter      bool          // calls router HandlerFunc in a goroutine if false
}

// sessionOption is the extra options for session.
type sessionOption struct {
	Packer        Packer
	Codec         Codec
	respQueueSize int
	asyncRouter   bool
}

// newSession creates a new session.
// Parameter conn is the TCP connection,
// opt includes packer, codec, and channel size.
// Returns a session pointer.
func newSession(conn net.Conn, opt *sessionOption) *session {
	return &session{
		id:               uuid.NewString(), // use uuid as default
		conn:             conn,
		closedC:          make(chan struct{}),
		afterCreateHookC: make(chan struct{}),
		afterCloseHookC:  make(chan struct{}),
		respStream:       make(chan Context, opt.respQueueSize),
		packer:           opt.Packer,
		codec:            opt.Codec,
		ctxPool:          sync.Pool{New: func() interface{} { return newContext() }},
		asyncRouter:      opt.asyncRouter,
	}
}

// ID returns the session's id.
func (s *session) ID() interface{} {
	return s.id
}

// SetID sets session id.
// Can be called in server.OnSessionCreate() callback.
func (s *session) SetID(id interface{}) {
	s.id = id
}

// Send pushes response message to respStream.
// Returns false if session is closed or ctx is done.
func (s *session) Send(ctx Context) (ok bool) {
	select {
	case <-ctx.Done():
		return false
	case <-s.closedC:
		return false
	case s.respStream <- ctx:
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
	s.closeOnce.Do(func() { close(s.closedC) })
}

// AfterCreateHook blocks until session's on-create hook triggered.
func (s *session) AfterCreateHook() <-chan struct{} {
	return s.afterCreateHookC
}

// AfterCloseHook blocks until session's on-close hook triggered.
func (s *session) AfterCloseHook() <-chan struct{} {
	return s.afterCloseHookC
}

// AllocateContext gets a Context from pool and reset all but session.
func (s *session) AllocateContext() Context {
	c := s.ctxPool.Get().(*routeContext)
	c.reset()
	c.SetSession(s)
	return c
}

// Conn returns the underlined connection instance.
func (s *session) Conn() net.Conn {
	return s.conn
}

// readInbound reads message packet from connection in a loop.
// And send unpacked message to reqQueue, which will be consumed in router.
// The loop breaks if errors occurred or the session is closed.
func (s *session) readInbound(router *Router, timeout time.Duration) {
	for {
		select {
		case <-s.closedC:
			return
		default:
		}
		if timeout > 0 {
			if err := s.conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
				_log.Errorf("session %s set read deadline err: %s", s.id, err)
				break
			}
		}
		reqMsg, err := s.packer.Unpack(s.conn)
		if err != nil {
			logMsg := fmt.Sprintf("session %s unpack inbound packet err: %s", s.id, err)
			if err == io.EOF {
				_log.Tracef(logMsg)
			} else {
				_log.Errorf(logMsg)
			}
			break
		}
		if reqMsg == nil {
			continue
		}

		if s.asyncRouter {
			go s.handleReq(router, reqMsg)
		} else {
			s.handleReq(router, reqMsg)
		}
	}
	_log.Tracef("session %s readInbound exit because of error", s.id)
	s.Close()
}

func (s *session) handleReq(router *Router, reqMsg *Message) {
	ctx := s.AllocateContext().SetRequestMessage(reqMsg)
	router.handleRequest(ctx)
	s.Send(ctx)
}

// writeOutbound fetches message from respStream channel and writes to TCP connection in a loop.
// Parameter writeTimeout specified the connection writing timeout.
// The loop breaks if errors occurred, or the session is closed.
func (s *session) writeOutbound(writeTimeout time.Duration) {
	for {
		var ctx Context
		select {
		case <-s.closedC:
			return
		case ctx = <-s.respStream:
		}

		outboundBytes, err := s.packResponse(ctx)
		if err != nil {
			_log.Errorf("session %s pack outbound message err: %s", s.id, err)
			continue
		}
		if outboundBytes == nil {
			continue
		}

		if writeTimeout > 0 {
			if err := s.conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
				_log.Errorf("session %s set write deadline err: %s", s.id, err)
				break
			}
		}

		if _, err := s.conn.Write(outboundBytes); err != nil {
			_log.Errorf("session %s conn write err: %s", s.id, err)
			break
		}
	}
	s.Close()
	_log.Tracef("session %s writeOutbound exit because of error", s.id)
}

func (s *session) packResponse(ctx Context) ([]byte, error) {
	defer s.ctxPool.Put(ctx)
	if ctx.Response() == nil {
		return nil, nil
	}
	return s.packer.Pack(ctx.Response())
}
