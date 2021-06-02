package session

import (
	"github.com/DarthPestilane/easytcp/packet"
)

type Session interface {
	ReadLoop()                            // read message from connection and do the unpack, decode stuff
	WriteLoop()                           // encode and pack message and write it to the connection
	RecvReq() <-chan *packet.Request      // fetch request from internal channel
	SendResp(resp *packet.Response) error // push resp into internal channel
	Close()
	WaitToClose() error
	ID() string               // get session id
	MsgPacker() packet.Packer // get message packer
	MsgCodec() packet.Codec   // get message codec
}
