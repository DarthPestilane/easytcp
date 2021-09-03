package easytcp

import (
	"fmt"
	"github.com/DarthPestilane/easytcp/internal/mock"
	"github.com/DarthPestilane/easytcp/message"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net"
	"sync"
	"testing"
	"time"
)

func TestNewTCPSession(t *testing.T) {
	var s *Session
	assert.NotPanics(t, func() {
		s = newSession(nil, &SessionOption{})
	})
	assert.NotNil(t, s)
	assert.NotNil(t, s.closed)
	assert.NotNil(t, s.respQueue)
	assert.NotNil(t, s.reqQueue)
}

func TestTCPSession_Close(t *testing.T) {
	sess := newSession(nil, &SessionOption{})
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sess.close() // goroutine safe
		}()
	}
	wg.Wait()
	_, ok := <-sess.closed
	assert.False(t, ok)
}

func TestTCPSession_ID(t *testing.T) {
	sess := newSession(nil, &SessionOption{})
	assert.NotEmpty(t, sess.id)
	assert.Equal(t, sess.ID(), sess.id)
}

func TestTCPSession_readLoop(t *testing.T) {
	t.Run("when connection set read timeout failed", func(t *testing.T) {
		p1, _ := net.Pipe()
		_ = p1.Close()
		sess := newSession(p1, &SessionOption{})
		go sess.readLoop(time.Millisecond)
		<-sess.closed
	})
	t.Run("when connection read timeout", func(t *testing.T) {
		p1, _ := net.Pipe()
		packer := &DefaultPacker{}
		sess := newSession(p1, &SessionOption{Packer: packer})
		go sess.readLoop(time.Millisecond * 10) // A timeout error is not fatal, we can keep going.
		time.Sleep(time.Millisecond * 12)
		_ = p1.Close()
		<-sess.closed
	})
	t.Run("when unpack message failed with error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		packer := mock.NewMockPacker(ctrl)
		packer.EXPECT().Unpack(gomock.Any()).Return(nil, fmt.Errorf("some error"))

		sess := newSession(nil, &SessionOption{Packer: packer, Codec: mock.NewMockCodec(ctrl)})
		go sess.readLoop(0)
		<-sess.closed
	})
	t.Run("when unpack message returns nil entry", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		packer := mock.NewMockPacker(ctrl)
		packer.EXPECT().Unpack(gomock.Any()).AnyTimes().Return(nil, nil)

		sess := newSession(nil, &SessionOption{Packer: packer, Codec: mock.NewMockCodec(ctrl)})
		go sess.readLoop(0)
		time.Sleep(time.Millisecond * 5)
		sess.close()
		<-sess.closed
	})
	t.Run("when session is closed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		packer := mock.NewMockPacker(ctrl)
		packer.EXPECT().Unpack(gomock.Any()).AnyTimes().Return(&message.Entry{ID: 1, Data: []byte("test")}, nil)

		sess := newSession(nil, &SessionOption{Packer: packer, ReadBufferSize: 1024})
		loopDone := make(chan struct{})
		go func() {
			sess.readLoop(0)
			close(loopDone)
		}()
		time.Sleep(time.Millisecond * 5)
		sess.close()
		<-loopDone
	})
	t.Run("when unpack message works fine", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		msg := &message.Entry{
			ID:   1,
			Data: []byte("unpacked"),
		}

		packer := mock.NewMockPacker(ctrl)
		packer.EXPECT().Unpack(gomock.Any()).AnyTimes().Return(msg, nil)

		sess := newSession(nil, &SessionOption{Packer: packer, Codec: mock.NewMockCodec(ctrl)})
		sess.reqQueue = make(chan *message.Entry) // no buffer
		readDone := make(chan struct{})
		go func() {
			sess.readLoop(0)
			close(readDone)
		}()
		req := <-sess.reqQueue
		time.Sleep(time.Millisecond * 5)
		sess.close() // close session once we fetched a req from channel
		assert.Equal(t, msg, req)
		<-readDone
	})
}

func TestTCPSession_SendResp(t *testing.T) {
	t.Run("when session is closed", func(t *testing.T) {
		entry := &message.Entry{
			ID:   1,
			Data: []byte("test"),
		}
		sess := newSession(nil, &SessionOption{})
		sess.close() // close session
		assert.Error(t, sess.SendResp(entry))
	})
	t.Run("when safelyPushRespQueue succeed", func(t *testing.T) {
		entry := &message.Entry{
			ID:   1,
			Data: []byte("test"),
		}

		sess := newSession(nil, &SessionOption{})
		sess.respQueue = make(chan *message.Entry) // no buffer
		go func() { <-sess.respQueue }()
		assert.NoError(t, sess.SendResp(entry))
		sess.close()
	})
}

func TestTCPSession_writeLoop(t *testing.T) {
	t.Run("when session is closed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		packer := mock.NewMockPacker(ctrl)
		packer.EXPECT().Pack(gomock.Any()).AnyTimes().Return(nil, nil)

		sess := newSession(nil, &SessionOption{Packer: packer, WriteBufferSize: 1024})
		sess.respQueue <- &message.Entry{}
		doneLoop := make(chan struct{})
		go func() {
			sess.writeLoop(0) // should stop looping and return
			close(doneLoop)
		}()
		time.Sleep(time.Millisecond * 5)
		sess.close()
		<-doneLoop
	})
	t.Run("when respQueue is closed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		packer := mock.NewMockPacker(ctrl)
		packer.EXPECT().Pack(gomock.Any()).AnyTimes().Return(nil, nil)

		sess := newSession(nil, &SessionOption{Packer: packer, WriteBufferSize: 1024})
		sess.respQueue <- &message.Entry{}
		doneLoop := make(chan struct{})
		go func() {
			sess.writeLoop(0) // should stop looping and return
			close(doneLoop)
		}()
		time.Sleep(time.Millisecond * 5)
		close(sess.respQueue)
		<-doneLoop
	})
	t.Run("when pack response message failed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		entry := &message.Entry{
			ID:   1,
			Data: []byte("test"),
		}
		packer := mock.NewMockPacker(ctrl)
		packer.EXPECT().Pack(gomock.Any()).Return(nil, fmt.Errorf("some err"))

		sess := newSession(nil, &SessionOption{Packer: packer})
		go func() { sess.respQueue <- entry }()
		time.Sleep(time.Microsecond * 15)
		go sess.writeLoop(0)
		time.Sleep(time.Millisecond * 15)
		sess.close() // should break the write loop
	})
	t.Run("when pack returns nil data", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		entry := &message.Entry{
			ID:   1,
			Data: []byte("test"),
		}
		packer := mock.NewMockPacker(ctrl)
		packer.EXPECT().Pack(gomock.Any()).Return(nil, nil)

		sess := newSession(nil, &SessionOption{Packer: packer, WriteBufferSize: 100})
		sess.respQueue <- entry // push to queue
		doneLoop := make(chan struct{})
		go func() {
			sess.writeLoop(0)
			close(doneLoop)
		}()
		time.Sleep(time.Millisecond * 5)
		sess.close() // should break the write loop
	})
	t.Run("when set write deadline failed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		entry := &message.Entry{
			ID:   1,
			Data: []byte("test"),
		}
		packer := mock.NewMockPacker(ctrl)
		packer.EXPECT().Pack(gomock.Any()).Return([]byte("pack succeed"), nil)

		p1, _ := net.Pipe()
		_ = p1.Close()
		sess := newSession(p1, &SessionOption{Packer: packer})
		go func() { sess.respQueue <- entry }()
		go sess.writeLoop(time.Millisecond * 10)
		_, ok := <-sess.closed
		assert.False(t, ok)
	})
	t.Run("when conn write timeout", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		entry := &message.Entry{
			ID:   1,
			Data: []byte("test"),
		}
		packer := mock.NewMockPacker(ctrl)
		packer.EXPECT().Pack(gomock.Any()).Return([]byte("pack succeed"), nil)

		p1, _ := net.Pipe()
		sess := newSession(p1, &SessionOption{Packer: packer})
		go func() { sess.respQueue <- entry }()
		go sess.writeLoop(time.Millisecond * 10)
		_, ok := <-sess.closed
		assert.False(t, ok)
		_ = p1.Close()
	})
	t.Run("when conn write failed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		entry := &message.Entry{
			ID:   1,
			Data: []byte("test"),
		}
		packer := mock.NewMockPacker(ctrl)
		packer.EXPECT().Pack(gomock.Any()).Return([]byte("pack succeed"), nil)

		p1, _ := net.Pipe()
		assert.NoError(t, p1.Close())
		sess := newSession(p1, &SessionOption{Packer: packer})
		go func() { sess.respQueue <- entry }()
		sess.writeLoop(0) // should stop looping and return
		_, ok := <-sess.closed
		assert.False(t, ok)
	})
}
