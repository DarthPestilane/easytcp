package router

import (
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/DarthPestilane/easytcp/session"
)

type Context struct {
	Session session.Session
	Request *packet.Request
}

func newContext(sess session.Session, req *packet.Request) *Context {
	return &Context{Session: sess, Request: req}
}

func (c *Context) MessageID() uint {
	return c.Request.ID
}

func (c *Context) MessageSize() uint {
	return c.Request.RawSize
}

func (c *Context) MessageRawData() []byte {
	return c.Request.RawData
}

func (c *Context) Bind(v interface{}) error {
	return c.Session.MsgCodec().Decode(c.Request.RawData, v)
}
