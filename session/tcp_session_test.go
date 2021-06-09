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

func TestNewTcp(t *testing.T) {
	var sess Session
	assert.NotPanics(t, func() {
		sess = NewTCP(nil, nil, nil)
	})
	assert.NotNil(t, sess)
	s, ok := sess.(*TCPSession)
	assert.True(t, ok)
	assert.NotNil(t, s.closed)
	assert.NotNil(t, s.ackQueue)
	assert.NotNil(t, s.reqQueue)
	assert.NotNil(t, s.log)
}

func TestTcpSession_Close(t *testing.T) {
	sess := NewTCP(nil, nil, nil)
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
	_, ok = <-sess.ackQueue
	assert.False(t, ok)
}

func TestTcpSession_ID(t *testing.T) {
	sess := NewTCP(nil, nil, nil)
	assert.NotEmpty(t, sess.id)
	assert.Equal(t, sess.ID(), sess.id)
}

func TestTcpSession_MsgCodec(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	codec := mock.NewMockCodec(ctrl)

	sess := NewTCP(nil, nil, codec)
	assert.NotNil(t, sess.msgCodec)
	assert.Equal(t, sess.msgCodec, codec)
	assert.Equal(t, sess.MsgCodec(), sess.msgCodec)
}

func TestTcpSession_ReadLoop(t *testing.T) {
	t.Run("when connection set read timeout failed", func(t *testing.T) {
		p1, _ := net.Pipe()
		_ = p1.Close()
		sess := NewTCP(p1, nil, nil)
		go sess.ReadLoop(time.Millisecond)
		<-sess.closed
	})
	t.Run("when connection read timeout", func(t *testing.T) {
		p1, _ := net.Pipe()
		packer := &packet.DefaultPacker{}
		sess := NewTCP(p1, packer, nil)
		go sess.ReadLoop(time.Millisecond * 10)
		<-sess.closed
		_ = p1.Close()
	})
	t.Run("when unpack message failed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		packer := mock.NewMockPacker(ctrl)
		packer.EXPECT().Unpack(gomock.Any()).Return(nil, fmt.Errorf("some err"))

		sess := NewTCP(nil, packer, mock.NewMockCodec(ctrl))
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

		sess := NewTCP(nil, packer, mock.NewMockCodec(ctrl))
		sess.reqQueue = make(chan *packet.Request) // no buffer
		readDone := make(chan struct{})
		go func() {
			sess.ReadLoop(0)
			readDone <- struct{}{}
		}()
		req := <-sess.reqQueue
		sess.Close() // close session once we fetched a req from channel
		expectReq := &packet.Request{
			ID:      msg.GetID(),
			RawSize: msg.GetSize(),
			RawData: msg.GetData(),
		}
		assert.Equal(t, expectReq, req)
		<-readDone
	})
}

func TestTcpSession_RecvReq(t *testing.T) {
	sess := NewTCP(nil, nil, nil)
	req := &packet.Request{}
	sess.reqQueue <- req
	reqRecv, ok := <-sess.RecvReq()
	assert.True(t, ok)
	assert.Equal(t, reqRecv, req)

	sess.Close()

	reqRecv, ok = <-sess.RecvReq()
	assert.False(t, ok)
	assert.Nil(t, reqRecv)
}

func TestTcpSession_SendResp(t *testing.T) {
	t.Run("when message encode failed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		codec := mock.NewMockCodec(ctrl)
		codec.EXPECT().Encode(gomock.Any()).Return(nil, fmt.Errorf("some err"))

		sess := NewTCP(nil, mock.NewMockPacker(ctrl), codec)
		closed, err := sess.SendResp(&packet.Response{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "encode response data err")
		assert.False(t, closed)
	})
	t.Run("when message pack failed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		codec := mock.NewMockCodec(ctrl)
		codec.EXPECT().Encode(gomock.Any()).Return([]byte("encoded"), nil)

		packer := mock.NewMockPacker(ctrl)
		packer.EXPECT().Pack(gomock.Any(), []byte("encoded")).Return(nil, fmt.Errorf("some err"))

		sess := NewTCP(nil, packer, codec)
		closed, err := sess.SendResp(&packet.Response{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pack response data err")
		assert.False(t, closed)
	})
	t.Run("when safelyPushAckQueue failed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		codec := mock.NewMockCodec(ctrl)
		codec.EXPECT().Encode(gomock.Any()).Return([]byte("encoded"), nil)

		packer := mock.NewMockPacker(ctrl)
		packer.EXPECT().Pack(gomock.Any(), []byte("encoded")).Return([]byte("packed"), nil)

		sess := NewTCP(nil, packer, codec)
		sess.Close()                                     // close channel
		closed, err := sess.SendResp(&packet.Response{}) // and then send resp
		assert.NoError(t, err)
		assert.True(t, closed)
	})
	t.Run("when safelyPushAckQueue succeed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		codec := mock.NewMockCodec(ctrl)
		codec.EXPECT().Encode(gomock.Any()).Return([]byte("encoded"), nil)

		packer := mock.NewMockPacker(ctrl)
		packer.EXPECT().Pack(gomock.Any(), []byte("encoded")).Return([]byte("packed"), nil)

		sess := NewTCP(nil, packer, codec)
		closed, err := sess.SendResp(&packet.Response{})
		sess.Close()
		assert.NoError(t, err)
		assert.False(t, closed)
	})
}

func TestTcpSession_WaitUntilClosed(t *testing.T) {
	sess := NewTCP(nil, nil, nil)
	go func() {
		sess.Close()
	}()
	sess.WaitUntilClosed()
	_, ok := <-sess.closed
	assert.False(t, ok)
}

func TestTcpSession_WriteLoop(t *testing.T) {
	t.Run("when session is already closed", func(t *testing.T) {
		sess := NewTCP(nil, nil, nil)
		sess.Close()
		sess.WriteLoop(0) // should stop looping and return
		_, ok := <-sess.closed
		assert.False(t, ok)
	})
	t.Run("when set write deadline failed", func(t *testing.T) {
		p1, _ := net.Pipe()
		_ = p1.Close()
		sess := NewTCP(p1, nil, nil)
		sess.ackQueue <- []byte("test")
		go sess.WriteLoop(time.Millisecond * 10)
		_, ok := <-sess.closed
		assert.False(t, ok)
	})
	t.Run("when conn write timeout", func(t *testing.T) {
		p1, _ := net.Pipe()
		sess := NewTCP(p1, nil, nil)
		sess.ackQueue <- []byte("test")
		go sess.WriteLoop(time.Millisecond * 10)
		_, ok := <-sess.closed
		assert.False(t, ok)
		_ = p1.Close()
	})
	t.Run("when conn write failed", func(t *testing.T) {
		p1, _ := net.Pipe()
		assert.NoError(t, p1.Close())
		sess := NewTCP(p1, nil, nil)
		sess.ackQueue <- []byte("test")
		sess.WriteLoop(0) // should stop looping and return
		_, ok := <-sess.closed
		assert.False(t, ok)
	})
	t.Run("it should return when session closed", func(t *testing.T) {
		p1, p2 := net.Pipe()
		go func() {
			buf := make([]byte, 8)
			_, err := p1.Read(buf)
			assert.NoError(t, err)
		}()
		sess := NewTCP(p2, nil, nil)
		sess.ackQueue = make(chan []byte) // no buffer
		writeDone := make(chan struct{})
		go func() {
			sess.WriteLoop(0)
			writeDone <- struct{}{}
		}()
		sess.ackQueue <- []byte("test")
		<-time.After(time.Millisecond * 100)
		sess.Close()
		_, ok := <-sess.closed
		assert.False(t, ok)
		<-writeDone
		assert.NoError(t, p1.Close())
		assert.NoError(t, p2.Close())
	})
}
