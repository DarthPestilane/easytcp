package easytcp

import (
	"bufio"
	"errors"
	"github.com/DarthPestilane/easytcp/logger"
	"github.com/DarthPestilane/easytcp/message"
	"net"
	"sync"
)

type Connection struct {
	net.Conn
	Closed      chan struct{} // to close(Closed)
	once        sync.Once
	handler     map[string]HandlerFunc
	msgBuffChan chan []byte
	msgChan     chan []byte
}

type ConnectionOption struct {
	BufferSize int
	Handler    map[string]HandlerFunc
}

func NewConnection(conn net.Conn, opt ConnectionOption) *Connection {
	return &Connection{
		Conn:        conn,
		handler:     opt.Handler,
		msgChan:     make(chan []byte),
		msgBuffChan: make(chan []byte, opt.BufferSize),
		Closed:      make(chan struct{}),
	}
}

func (c *Connection) NetConn() net.Conn {
	return c.Conn
}

func (c *Connection) Send(routePath string, data []byte) error {
	if c.isClosed() {
		return errors.New("connection has already been closed")
	}
	msg := message.AddHead(routePath, data)
	c.msgChan <- msg
	return nil
}

func (c *Connection) SendBuffer(routePath string, data []byte) error {
	if c.isClosed() {
		return errors.New("connection has already been closed")
	}
	msg := message.AddHead(routePath, data)
	c.msgBuffChan <- msg
	return nil
}

func (c *Connection) KeepReading() {
	defer c.Close()
	for {
		head, body, err := c.ReadMessage()
		if err != nil {
			logger.Default.Errorf("read connection %s failed: %s", c.RemoteAddr(), err)
			return
		}
		handlerCtx := NewContext()
		handlerCtx.setConn(c).setBody(body).setLength(head.Length).setRoutePath(head.RoutePath)
		if c.handler != nil && c.handler[head.RoutePath] != nil {
			go c.handler[head.RoutePath](handlerCtx)
		}
	}
}

func (c *Connection) ReadMessage() (head *message.Head, body []byte, err error) {
	connReader := bufio.NewReader(c.Conn)
	headByte, err := connReader.ReadBytes('|')
	if err != nil {
		return nil, nil, err
	}
	logger.Default.Debugf("msg head: %s", headByte)
	head, err = message.ExtractHead(headByte)
	if err != nil {
		return nil, nil, err
	}
	body = make([]byte, head.Length)
	n, err := connReader.Read(body)
	if err != nil {
		return nil, nil, err
	}
	return head, body[:n], nil
}

func (c *Connection) KeepWriting() {
	for {
		select {
		case <-c.Closed:
			return
		case msg := <-c.msgChan:
			n, err := c.Conn.Write(msg)
			if err != nil {
				logger.Default.Errorf("send data failed: %s", err)
				break
			}
			logger.Default.Debugf("send %d bytes data", n)
		case msg := <-c.msgBuffChan:
			n, err := c.Conn.Write(msg)
			if err != nil {
				logger.Default.Errorf("send bufferd data failed: %s", err)
				break
			}
			logger.Default.Debugf("send %d bytes bufferd data", n)
		}
	}
}

func (c *Connection) Close() {
	c.once.Do(func() {
		_ = c.Conn.Close()
		close(c.Closed)
	})
}

func (c *Connection) isClosed() bool {
	select {
	case <-c.Closed:
		return true
	default:
		return false
	}
}
