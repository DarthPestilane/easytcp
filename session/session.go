package session

import (
	"github.com/DarthPestilane/easytcp/packet"
)

type Session interface {
	RecvReq() <-chan *packet.Request      // fetch request from internal channel
	SendResp(resp *packet.Response) error // push resp into internal channel
	ID() string                           // get session id
	MsgPacker() packet.Packer             // get message packer
	MsgCodec() packet.Codec               // get message codec
}
