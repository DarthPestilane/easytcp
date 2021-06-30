package session

import (
	"fmt"
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/DarthPestilane/easytcp/packet/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

func TestNewUDPSession(t *testing.T) {
	var sess Session
	assert.NotPanics(t, func() {
		sess = NewUDPSession(nil, nil, nil, nil)
	})
	assert.NotNil(t, sess)
	s, ok := sess.(*UDPSession)
	assert.True(t, ok)
	assert.NotNil(t, s.closed)
	assert.NotNil(t, s.respQueue)
	assert.NotNil(t, s.reqQueue)
	assert.NotNil(t, s.log)
}

func TestUDPSession_Close(t *testing.T) {
	sess := NewUDPSession(nil, nil, nil, nil)
	sess.Close()
	var ok bool
	_, ok = <-sess.closed
	assert.False(t, ok)
	_, ok = <-sess.reqQueue
	assert.False(t, ok)
	_, ok = <-sess.respQueue
	assert.False(t, ok)
}

func TestUDPSession_ID(t *testing.T) {
	sess := NewUDPSession(nil, nil, nil, nil)
	assert.NotEmpty(t, sess.ID())
	assert.Equal(t, sess.id, sess.ID())
}

func TestUDPSession_MsgCodec(t *testing.T) {
	sess := NewUDPSession(nil, nil, nil, &packet.JsonCodec{})
	assert.NotNil(t, sess.MsgCodec())
	assert.Equal(t, sess.msgCodec, &packet.JsonCodec{})
	assert.Equal(t, sess.msgCodec, sess.MsgCodec())
}

func TestUDPSession_ReadIncomingMsg(t *testing.T) {
	t.Run("when packer unpack failed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		packer := mock.NewMockPacker(ctrl)
		packer.EXPECT().Unpack(gomock.Any()).Return(nil, fmt.Errorf("some err"))

		sess := NewUDPSession(nil, nil, packer, nil)
		go func() { <-sess.reqQueue }()
		assert.Error(t, sess.ReadIncomingMsg([]byte("test")))
		sess.Close()
	})
	t.Run("when reqQueue closed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		msg := &packet.MessageEntry{
			ID:   1,
			Data: []byte("test"),
		}
		packer := mock.NewMockPacker(ctrl)
		packer.EXPECT().Unpack(gomock.Any()).Return(msg, nil)

		sess := NewUDPSession(nil, nil, packer, nil)
		sess.Close() // close first
		assert.NoError(t, sess.ReadIncomingMsg([]byte("test")))
	})
	t.Run("when succeed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		msg := &packet.MessageEntry{
			ID:   1,
			Data: []byte("test"),
		}
		packer := mock.NewMockPacker(ctrl)
		packer.EXPECT().Unpack(gomock.Any()).Return(msg, nil)

		sess := NewUDPSession(nil, nil, packer, nil)
		go func() { <-sess.reqQueue }()
		assert.NoError(t, sess.ReadIncomingMsg([]byte("test")))
		sess.Close()
	})
}

func TestUDPSession_RecvReq(t *testing.T) {
	msg := &packet.MessageEntry{
		ID:   1,
		Data: []byte("test"),
	}

	sess := NewUDPSession(nil, nil, nil, nil)
	go func() { sess.reqQueue <- msg }()
	req, ok := <-sess.RecvReq()
	assert.True(t, ok)
	assert.Equal(t, req, msg)
	sess.Close()
	_, ok = <-sess.RecvReq()
	assert.False(t, ok)
}

func TestUDPSession_SendResp(t *testing.T) {
	t.Run("when respQueue closed (session's closed)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		message := &packet.MessageEntry{
			ID:   1,
			Data: []byte("test"),
		}
		codec := mock.NewMockCodec(ctrl)
		packer := mock.NewMockPacker(ctrl)

		sess := NewUDPSession(nil, nil, packer, codec)
		sess.Close()
		assert.Error(t, sess.SendResp(message))
	})
	t.Run("when succeed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		message := &packet.MessageEntry{
			ID:   1,
			Data: []byte("test"),
		}
		codec := mock.NewMockCodec(ctrl)
		packer := mock.NewMockPacker(ctrl)

		sess := NewUDPSession(nil, nil, packer, codec)
		go func() { <-sess.respQueue }()
		assert.NoError(t, sess.SendResp(message))
		sess.Close()
	})
}

func TestUDPSession_Write(t *testing.T) {
	t.Run("when done closed", func(t *testing.T) {
		sess := NewUDPSession(nil, nil, nil, nil)
		done := make(chan struct{})
		go func() { close(sess.respQueue) }()
		sess.Write(done)
	})
	t.Run("when respQueue closed", func(t *testing.T) {
		sess := NewUDPSession(nil, nil, nil, nil)
		done := make(chan struct{})
		go func() { close(done) }()
		sess.Write(done)
	})
	t.Run("when pack response message failed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		message := &packet.MessageEntry{
			ID:   1,
			Data: []byte("test"),
		}
		packer := mock.NewMockPacker(ctrl)
		packer.EXPECT().Pack(gomock.Any()).Return(nil, fmt.Errorf("some err"))

		done := make(chan struct{})
		sess := NewUDPSession(nil, nil, packer, nil)
		go func() { sess.respQueue <- message }()
		sess.Write(done)
		assert.True(t, true)
	})
	t.Run("when conn write failed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		message := &packet.MessageEntry{
			ID:   1,
			Data: []byte("test"),
		}
		packer := mock.NewMockPacker(ctrl)
		packer.EXPECT().Pack(gomock.Any()).Return([]byte("pack succeed"), nil)

		addr, err := net.ResolveUDPAddr("udp", "localhost:0")
		assert.NoError(t, err)
		conn, err := net.ListenUDP("udp", addr)
		assert.NoError(t, err)

		sess := NewUDPSession(conn, nil, packer, nil)
		done := make(chan struct{})
		go func() { sess.respQueue <- message }()
		sess.Write(done)
	})
}
