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
	Closed      chan struct{} // to close(Closed)
	mu          sync.Mutex
	server      *Server
	msgBuffChan chan []byte
	msgChan     chan []byte
}

func NewConnection(conn net.Conn, server *Server) *Connection {
	return &Connection{
		Conn:        conn,
		server:      server,
		msgChan:     make(chan []byte),
		msgBuffChan: make(chan []byte, 1024),
		Closed:      make(chan struct{}),
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
		case <-c.Closed:
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
	if c.alreadyClosed() {
		return
	}
	_ = c.Conn.Close()
	close(c.Closed)
}

func (c *Connection) alreadyClosed() bool {
	select {
	case <-c.Closed:
		return true
	default:
		return false
	}
}
