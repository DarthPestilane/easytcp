package easytcp

import (
	"errors"
	"fmt"
	"net"
	"time"
)

type Client struct {
	Conn net.Conn

	// Packer is the message packer, will be passed to session.
	Packer Packer

	// Codec is the message codec, will be passed to session.
	Codec Codec

	// OnSessionCreate is an event hook, will be invoked when session's created.
	OnSessionCreate func(sess Session)

	// OnSessionClose is an event hook, will be invoked when session's closed.
	OnSessionClose func(sess Session)

	socketReadBufferSize  int
	socketWriteBufferSize int
	socketSendDelay       bool
	readTimeout           time.Duration
	writeTimeout          time.Duration
	respQueueSize         int
	router                *Router
	printRoutes           bool
	stopped               chan struct{}
	writeAttemptTimes     int
	asyncRouter           bool

	Sess       *session
	notifyChan chan interface{}
}

type ClientOption struct {
	ServerOption
	NotifyChan chan interface{}
}

func NewClient(opt *ClientOption) *Client {
	if opt.Packer == nil {
		opt.Packer = NewDefaultPacker()
	}
	if opt.RespQueueSize < 0 {
		opt.RespQueueSize = DefaultRespQueueSize
	}
	if opt.WriteAttemptTimes <= 0 {
		opt.WriteAttemptTimes = DefaultWriteAttemptTimes
	}
	return &Client{
		socketReadBufferSize:  opt.SocketReadBufferSize,
		socketWriteBufferSize: opt.SocketWriteBufferSize,
		socketSendDelay:       opt.SocketSendDelay,
		respQueueSize:         opt.RespQueueSize,
		readTimeout:           opt.ReadTimeout,
		writeTimeout:          opt.WriteTimeout,
		Packer:                opt.Packer,
		Codec:                 opt.Codec,
		printRoutes:           !opt.DoNotPrintRoutes,
		router:                newRouter(),
		stopped:               make(chan struct{}),
		writeAttemptTimes:     opt.WriteAttemptTimes,
		asyncRouter:           opt.AsyncRouter,
		notifyChan:            opt.NotifyChan,
	}
}

func (c *Client) Run(addr string) error {
	dial, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	c.Conn = dial
	go c.handleConn(c.Conn)
	return nil
}

func (c *Client) handleConn(conn net.Conn) {
	defer conn.Close() // nolint

	sess := newSession(conn, &sessionOption{
		Packer:        c.Packer,
		Codec:         c.Codec,
		respQueueSize: c.respQueueSize,
		asyncRouter:   c.asyncRouter,
		notifyChan:    c.notifyChan,
	})
	if c.OnSessionCreate != nil {
		c.OnSessionCreate(sess)
	}
	close(sess.afterCreateHook)
	c.Sess = sess

	go sess.readInbound(c.router, c.readTimeout)               // start reading message packet from connection.
	go sess.writeOutbound(c.writeTimeout, c.writeAttemptTimes) // start writing message packet to connection.

	select {
	case <-sess.closed: // wait for session finished.
	case <-c.stopped: // or the server is stopped.
	}

	if c.OnSessionClose != nil {
		c.OnSessionClose(sess)
	}
	close(sess.afterCloseHook)
}

// Stop stops server. Closing Listener and all connections.
func (c *Client) Stop() error {
	close(c.stopped)
	return c.Conn.Close()
}

// AddRoute registers message handler and middlewares to the router.
func (c *Client) AddRoute(msgID interface{}, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	c.router.register(msgID, handler, middlewares...)
}

// Use registers global middlewares to the router.
func (c *Client) Use(middlewares ...MiddlewareFunc) {
	c.router.registerMiddleware(middlewares...)
}

// NotFoundHandler sets the not-found handler for router.
func (c *Client) NotFoundHandler(handler HandlerFunc) {
	c.router.setNotFoundHandler(handler)
}

func (c *Client) IsStopped() bool {
	select {
	case <-c.stopped:
		return true
	default:
		return false
	}
}

func (c *Client) Send(id, v interface{}) error {
	if c.Codec == nil {
		return errors.New("codec is nil")
	}
	data, err := c.Codec.Encode(v)
	if err != nil {
		return fmt.Errorf("encode message failed: %v", err)
	}
	return c.SendMsg(NewMessage(id, data))
}

func (c *Client) SendMsg(msg *Message) error {
	ctx := c.Sess.AllocateContext().SetResponseMessage(msg)
	if c.IsStopped() {
		return errors.New("client is stopped")
	}
	if ok := c.Sess.Send(ctx); !ok {
		return errors.New("send message failed")
	}
	return nil
}
