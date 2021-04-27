package easytcp

import (
	"bufio"
	"errors"
	"github.com/DarthPestilane/easytcp/message"
	"github.com/sirupsen/logrus"
	"net"
	"sync"
)

type Connection struct {
	net.Conn
	mu          sync.Mutex
	server      *Server
	msgBuffChan chan []byte
	msgChan     chan []byte
	isClosed    bool
	closeChan   chan struct{} // to close(closeChan)
}

func NewConnection(conn net.Conn, server *Server) *Connection {
	return &Connection{
		Conn:        conn,
		server:      server,
		msgChan:     make(chan []byte),
		msgBuffChan: make(chan []byte, 1024),
		closeChan:   make(chan struct{}),
	}
}

func (c *Connection) NetConn() net.Conn {
	return c.Conn
}

func (c *Connection) Send(routePath string, data []byte) error {
	if c.alreadyClosed() {
		return errors.New("connection has already been closed")
	}
	msg := message.AddHead(routePath, data)
	c.msgChan <- msg
	return nil
}

func (c *Connection) SendBuffer(routePath string, data []byte) error {
	if c.alreadyClosed() {
		return errors.New("connection has already been closed")
	}
	msg := message.AddHead(routePath, data)
	c.msgBuffChan <- msg
	return nil
}

func (c *Connection) KeepReading() {
	defer c.Close()
	for {
		connReader := bufio.NewReader(c)
		head, err := connReader.ReadBytes('|')
		if err != nil {
			logrus.Errorf("read conn %s message head failed: %s", c.RemoteAddr(), err)
			return
		}
		logrus.Debugf("msg head: %s", head)
		headStruct, err := message.ExtractHead(head)
		if err != nil {
			logrus.Error(err)
			continue
		}
		body := make([]byte, headStruct.Length)
		n, err := connReader.Read(body)
		if err != nil {
			logrus.Errorf("read conn %s message body failed: %s", c.RemoteAddr(), err)
			return
		}

		handlerCtx := NewContext()
		handlerCtx.setConn(c).setBody(body[:n]).setLength(n).setRoutePath(headStruct.RoutePath)
		go c.server.handler[headStruct.RoutePath](handlerCtx)
	}
}

func (c *Connection) KeepWriting() {
	for {
		select {
		case <-c.closeChan:
			return
		case msg := <-c.msgChan:
			n, err := c.Conn.Write(msg)
			if err != nil {
				logrus.Errorf("send data failed: %s", err)
				break
			}
			logrus.Infof("send %d bytes data", n)
		case msg := <-c.msgBuffChan:
			n, err := c.Conn.Write(msg)
			if err != nil {
				logrus.Errorf("send bufferd data failed: %s", err)
				break
			}
			logrus.Infof("send %d bytes bufferd data", n)
		}
	}
}

func (c *Connection) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.isClosed {
		return
	}
	if err := c.Conn.Close(); err != nil {
		logrus.Errorf("close connection failed: %s", err)
		return
	}
	c.isClosed = true
	close(c.closeChan)
}

func (c *Connection) alreadyClosed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.isClosed
}
