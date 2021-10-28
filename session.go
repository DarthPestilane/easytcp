package easytcp

import (
	"fmt"
	"github.com/google/uuid"
	"net"
	"sync"
	"time"
)

type ISession interface {
	ID() string
	SendResp(ctx *Context) error
	Close()
}

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

// sessionOption is the extra options for Session.
type sessionOption struct {
	Packer        Packer
	Codec         Codec
	respQueueSize int
}

// newSession creates a new Session.
// Parameter conn is the TCP connection,
// opt includes packer, codec, and channel size.
// Returns a Session pointer.
func newSession(conn net.Conn, opt *sessionOption) *Session {
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
	s.respQueue <- ctx
	return
}

// Close closes the session, but doesn't close the connection.
// The connection will be closed in the server once the session's closed.
func (s *Session) Close() {
	defer func() { _ = recover() }()
	close(s.closed)
	close(s.respQueue)
}

// readInbound reads message packet from connection in a loop.
// And send unpacked message to reqQueue, which will be consumed in router.
// The loop breaks if errors occurred or the session is closed.
func (s *Session) readInbound(router *Router, timeout time.Duration) {
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
			ctx := s.ctxPool.Get().(*Context)
			ctx.reset(s, reqEntry)
			if err := router.handleRequest(ctx); err != nil {
				Log.Errorf("handle request err: %s", err)
			}
			if err := s.SendResp(ctx); err != nil {
				Log.Errorf("send resp context err: %s", err)
			}
		}()
	}
	Log.Tracef("session %s readInbound exit because of error", s.id)
	s.Close()
}

// writeOutbound fetches message from respQueue channel and writes to TCP connection in a loop.
// Parameter writeTimeout specified the connection writing timeout.g
// The loop breaks if errors occurred, or the session is closed.
func (s *Session) writeOutbound(writeTimeout time.Duration, attemptTimes int) {
	for {
		ctx, ok := <-s.respQueue
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

func (s *Session) attemptConnWrite(outboundMsg []byte, attemptTimes int) (err error) {
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

func (s *Session) pack(ctx *Context) ([]byte, error) {
	defer s.ctxPool.Put(ctx)
	if ctx.respEntry == nil {
		return nil, nil
	}
	return s.packer.Pack(ctx.respEntry)
}
