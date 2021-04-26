package tcp_demo

import (
	"bufio"
	"demo/tcp_demo/util"
	"demo/tcp_demo/util/message"
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"sync"
	"time"
)

type Client struct {
	DialAddr string
	DialPort int
	Conn     net.Conn
	mu       sync.Mutex
	handler  map[string]HandlerFunc
}

func NewClient(dialAddr string, dialPort int) *Client {
	return &Client{
		DialAddr: dialAddr,
		DialPort: dialPort,
		handler:  make(map[string]HandlerFunc),
	}
}

func (c *Client) Dial(timeout time.Duration) error {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", c.DialAddr, c.DialPort), timeout)
	if err != nil {
		return err
	}
	c.Conn = conn
	return nil
}

func (c *Client) AddRoute(routePath string, fn HandlerFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handler[routePath] = fn
}

func (c *Client) Send(routePath string, b []byte) (int, error) {
	msg := message.AddHead(routePath, b)
	msg = append(msg, '\n')
	return c.Conn.Write(msg)
}

func (c *Client) SendIn(routePath string, b []byte, duration time.Duration) (int, error) {
	msg := message.AddHead(routePath, b)
	msg = append(msg, '\n')

	if err := c.Conn.SetWriteDeadline(time.Now().Add(duration)); err != nil {
		return 0, err
	}
	defer c.Conn.SetWriteDeadline(time.Time{})
	return c.Conn.Write(msg)
}

func (c *Client) StartReading() {
	conn := c.Conn
	for {
		connReader := bufio.NewReader(conn)
		head, err := connReader.ReadBytes('|')
		if err != nil {
			if util.IsEOF(err) {
				logrus.Errorf("server disconnected!! %s:%d", c.DialAddr, c.DialPort)
				return
			}
			logrus.Errorf("read head error: %s", err)
			continue
		}
		headStruct, err := message.ExtractHead(head)
		if err != nil {
			logrus.Errorf("invalid message head: %s", err)
			continue
		}
		body := make([]byte, headStruct.Length)
		n, err := connReader.Read(body)
		if err != nil {
			if util.IsEOF(err) {
				logrus.Errorf("server disconnected!! %s:%d", c.DialAddr, c.DialPort)
				return
			}
			logrus.Errorf("read body error: %s", err)
			continue
		}
		ctx := NewContext()
		ctx.setConn(conn).setBody(body[:n])
		c.handler[headStruct.RoutePath](ctx)
	}
}
