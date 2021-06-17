package session

import (
	"fmt"
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/DarthPestilane/easytcp/packet/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net"
	"sync"
	"testing"
	"time"
)

func TestNewTCPSession(t *testing.T) {
	var sess Session
	assert.NotPanics(t, func() {
		sess = NewTCPSession(nil, nil, nil)
	})
	assert.NotNil(t, sess)
	s, ok := sess.(*TCPSession)
	assert.True(t, ok)
	assert.NotNil(t, s.closed)
	assert.NotNil(t, s.respQueue)
	assert.NotNil(t, s.reqQueue)
	assert.NotNil(t, s.log)
}

func TestTCPSession_Close(t *testing.T) {
	sess := NewTCPSession(nil, nil, nil)
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sess.Close() // goroutine safe
		}()
	}
	wg.Wait()
	_, ok := <-sess.closed
	assert.False(t, ok)
	_, ok = <-sess.reqQueue
	assert.False(t, ok)
	_, ok = <-sess.respQueue
	assert.False(t, ok)
}

func TestTCPSession_ID(t *testing.T) {
	sess := NewTCPSession(nil, nil, nil)
	assert.NotEmpty(t, sess.id)
	assert.Equal(t, sess.ID(), sess.id)
}

func TestTCPSession_MsgCodec(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	codec := mock.NewMockCodec(ctrl)

	sess := NewTCPSession(nil, nil, codec)
	assert.NotNil(t, sess.msgCodec)
	assert.Equal(t, sess.msgCodec, codec)
	assert.Equal(t, sess.MsgCodec(), sess.msgCodec)
}

func TestTCPSession_ReadLoop(t *testing.T) {
	t.Run("when connection set read timeout failed", func(t *testing.T) {
		p1, _ := net.Pipe()
		_ = p1.Close()
		sess := NewTCPSession(p1, nil, nil)
		go sess.ReadLoop(time.Millisecond)
		<-sess.closed
	})
	t.Run("when connection read timeout", func(t *testing.T) {
		p1, _ := net.Pipe()
		packer := &packet.DefaultPacker{}
		sess := NewTCPSession(p1, packer, nil)
		go sess.ReadLoop(time.Millisecond * 10)
		<-sess.closed
		_ = p1.Close()
	})
	t.Run("when unpack message failed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		packer := mock.NewMockPacker(ctrl)
		packer.EXPECT().Unpack(gomock.Any()).Return(nil, fmt.Errorf("some err"))

		sess := NewTCPSession(nil, packer, mock.NewMockCodec(ctrl))
		go sess.ReadLoop(0)
		<-sess.closed
	})
	t.Run("when unpack message works fine", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		msg := mock.NewMockMessage(ctrl)
		msg.EXPECT().GetID().AnyTimes().Return(uint(1))
		msg.EXPECT().GetSize().AnyTimes().Return(uint(len("unpacked")))
		msg.EXPECT().GetData().AnyTimes().Return([]byte("unpacked"))

		packer := mock.NewMockPacker(ctrl)
		packer.EXPECT().Unpack(gomock.Any()).AnyTimes().Return(msg, nil)

		sess := NewTCPSession(nil, packer, mock.NewMockCodec(ctrl))
		sess.reqQueue = make(chan packet.Message) // no buffer
		readDone := make(chan struct{})
		go func() {
			sess.ReadLoop(0)
			readDone <- struct{}{}
		}()
		req := <-sess.reqQueue
		sess.Close() // close session once we fetched a req from channel
		assert.Equal(t, msg, req)
		<-readDone
	})
}

func TestTCPSession_RecvReq(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	msg := mock.NewMockMessage(ctrl)

	sess := NewTCPSession(nil, nil, nil)
	sess.reqQueue <- msg
	reqRecv, ok := <-sess.RecvReq()
	assert.True(t, ok)
	assert.Equal(t, reqRecv, msg)

	sess.Close()

	reqRecv, ok = <-sess.RecvReq()
	assert.False(t, ok)
	assert.Nil(t, reqRecv)
}

func TestTCPSession_SendResp(t *testing.T) {
	t.Run("when safelyPushRespQueue failed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		message := mock.NewMockMessage(ctrl)
		codec := mock.NewMockCodec(ctrl)
		packer := mock.NewMockPacker(ctrl)

		sess := NewTCPSession(nil, packer, codec)
		sess.Close()                            // close channel
		assert.Error(t, sess.SendResp(message)) // and then send resp
	})
	t.Run("when safelyPushRespQueue succeed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		message := mock.NewMockMessage(ctrl)
		codec := mock.NewMockCodec(ctrl)
		packer := mock.NewMockPacker(ctrl)

		sess := NewTCPSession(nil, packer, codec)
		sess.respQueue = make(chan packet.Message) // no buffer
		go func() { <-sess.respQueue }()
		assert.NoError(t, sess.SendResp(message))
		sess.Close()
	})
}

func TestTCPSession_WaitUntilClosed(t *testing.T) {
	sess := NewTCPSession(nil, nil, nil)
	go func() {
		sess.Close()
	}()
	sess.WaitUntilClosed()
	_, ok := <-sess.closed
	assert.False(t, ok)
}

func TestTCPSession_WriteLoop(t *testing.T) {
	t.Run("when session is already closed", func(t *testing.T) {
		sess := NewTCPSession(nil, nil, nil)
		sess.Close()
		sess.WriteLoop(0) // should stop looping and return
		_, ok := <-sess.closed
		assert.False(t, ok)
	})
	t.Run("when pack response message failed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		message := mock.NewMockMessage(ctrl)
		packer := mock.NewMockPacker(ctrl)
		packer.EXPECT().Pack(gomock.Any()).Return(nil, fmt.Errorf("some err"))

		sess := NewTCPSession(nil, packer, nil)
		go sess.WriteLoop(0)
		time.Sleep(time.Millisecond * 5)
		sess.respQueue <- message
		time.Sleep(time.Millisecond * 5)
		sess.Close() // should break the write loop
		assert.True(t, true)
	})
	t.Run("when set write deadline failed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		message := mock.NewMockMessage(ctrl)
		packer := mock.NewMockPacker(ctrl)
		packer.EXPECT().Pack(gomock.Any()).Return([]byte("pack succeed"), nil)

		p1, _ := net.Pipe()
		_ = p1.Close()
		sess := NewTCPSession(p1, packer, nil)
		sess.respQueue <- message
		go sess.WriteLoop(time.Millisecond * 10)
		_, ok := <-sess.closed
		assert.False(t, ok)
	})
	t.Run("when conn write timeout", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		message := mock.NewMockMessage(ctrl)
		packer := mock.NewMockPacker(ctrl)
		packer.EXPECT().Pack(gomock.Any()).Return([]byte("pack succeed"), nil)

		p1, _ := net.Pipe()
		sess := NewTCPSession(p1, packer, nil)
		sess.respQueue <- message
		go sess.WriteLoop(time.Millisecond * 10)
		_, ok := <-sess.closed
		assert.False(t, ok)
		_ = p1.Close()
	})
	t.Run("when conn write failed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		message := mock.NewMockMessage(ctrl)
		packer := mock.NewMockPacker(ctrl)
		packer.EXPECT().Pack(gomock.Any()).Return([]byte("pack succeed"), nil)

		p1, _ := net.Pipe()
		assert.NoError(t, p1.Close())
		sess := NewTCPSession(p1, packer, nil)
		sess.respQueue <- message
		sess.WriteLoop(0) // should stop looping and return
		_, ok := <-sess.closed
		assert.False(t, ok)
	})
}
