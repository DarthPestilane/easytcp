package session

import (
	"github.com/DarthPestilane/easytcp/packet"
)

//go:generate mockgen -destination mock/session_mock.go -package mock . Session

// Session handles message receiving and sending
type Session interface {
	ID() string                                              // get session id
	MsgCodec() packet.Codec                                  // get message codec
	RecvReq() <-chan *packet.Request                         // fetch request from internal channel
	SendResp(resp *packet.Response) (closed bool, err error) // push resp into internal channel
	Close()                                                  // close current session, exit corresponding goroutines
}
