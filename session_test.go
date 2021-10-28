package easytcp

import (
	"fmt"
	"github.com/DarthPestilane/easytcp/internal/mock"
	"github.com/DarthPestilane/easytcp/message"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"io"
	"net"
	"sync"
	"testing"
	"time"
)

func TestNewTCPSession(t *testing.T) {
	var s *session
	assert.NotPanics(t, func() {
		s = newSession(nil, &sessionOption{})
	})
	assert.NotNil(t, s)
	assert.NotNil(t, s.closed)
	assert.NotNil(t, s.respQueue)
}

func TestTCPSession_close(t *testing.T) {
	sess := newSession(nil, &sessionOption{})
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
}

func TestTCPSession_ID(t *testing.T) {
	sess := newSession(nil, &sessionOption{})
	assert.NotEmpty(t, sess.id)
	assert.Equal(t, sess.ID(), sess.id)
}

func TestTCPSession_readInbound(t *testing.T) {
	t.Run("when connection set read timeout failed", func(t *testing.T) {
		p1, _ := net.Pipe()
		_ = p1.Close()
		sess := newSession(p1, &sessionOption{})
		go sess.readInbound(nil, time.Millisecond)
		<-sess.closed
	})
	t.Run("when connection read timeout", func(t *testing.T) {
		p1, _ := net.Pipe()
		packer := &DefaultPacker{}
		sess := newSession(p1, &sessionOption{Packer: packer})
		done := make(chan struct{})
		go func() {
			sess.readInbound(nil, time.Millisecond*10)
			close(done)
		}()
		<-done
	})
	t.Run("when unpack message failed with error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		packer := mock.NewMockPacker(ctrl)
		packer.EXPECT().Unpack(gomock.Any()).Return(nil, fmt.Errorf("some error"))

		sess := newSession(nil, &sessionOption{Packer: packer, Codec: mock.NewMockCodec(ctrl)})
		done := make(chan struct{})
		go func() {
			sess.readInbound(nil, 0)
			close(done)
		}()
		<-done
		<-sess.closed
	})
	t.Run("when unpack message returns nil entry", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		first := true
		packer := mock.NewMockPacker(ctrl)
		packer.EXPECT().Unpack(gomock.Any()).Times(2).DoAndReturn(func(_ io.Reader) (*message.Entry, error) {
			if first {
				first = false
				return nil, nil
			} else {
				return nil, fmt.Errorf("unpack error")
			}
		})

		sess := newSession(nil, &sessionOption{Packer: packer, Codec: mock.NewMockCodec(ctrl)})
		done := make(chan struct{})
		go func() {
			sess.readInbound(nil, 0)
			close(done)
		}()
		<-done
	})
	t.Run("when send response failed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		packer := mock.NewMockPacker(ctrl)
		packer.EXPECT().Unpack(gomock.Any()).AnyTimes().Return(&message.Entry{ID: 1, Data: []byte("test")}, nil)

		r := newRouter()
		r.register(1, func(ctx *Context) error {
			ctx.session.Close()
			return fmt.Errorf("route error")
		})

		sess := newSession(nil, &sessionOption{Packer: packer, respQueueSize: 10})
		loopDone := make(chan struct{})
		go func() {
			sess.readInbound(r, 0)
			close(loopDone)
		}()
		<-loopDone
	})
	t.Run("when unpack message works fine", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		first := true
		packer := mock.NewMockPacker(ctrl)
		packer.EXPECT().Unpack(gomock.Any()).Times(2).DoAndReturn(func(_ io.Reader) (*message.Entry, error) {
			if first {
				first = false
				return &message.Entry{ID: 1, Data: []byte("unpack ok")}, nil
			} else {
				return nil, fmt.Errorf("unpack error")
			}
		})

		r := newRouter()
		r.register(1, func(ctx *Context) error {
			return ctx.Response(2, []byte("ok"))
		})

		sess := newSession(nil, &sessionOption{Packer: packer, Codec: nil, respQueueSize: 10})
		readDone := make(chan struct{})
		go func() {
			sess.readInbound(r, 0)
			close(readDone)
		}()
		<-readDone
	})
}

func TestTCPSession_SendResp(t *testing.T) {
	t.Run("when session is closed", func(t *testing.T) {
		entry := &message.Entry{
			ID:   1,
			Data: []byte("test"),
		}
		sess := newSession(nil, &sessionOption{})
		sess.Close() // close session
		assert.Error(t, sess.SendResp(&Context{respEntry: entry}))
	})
	t.Run("when session respQueue is closed", func(t *testing.T) {
		sess := newSession(nil, &sessionOption{})
		close(sess.respQueue)
		assert.Error(t, sess.SendResp(nil))
	})
	t.Run("when send succeed", func(t *testing.T) {
		entry := &message.Entry{
			ID:   1,
			Data: []byte("test"),
		}

		sess := newSession(nil, &sessionOption{})
		sess.respQueue = make(chan *Context) // no buffer
		go func() { <-sess.respQueue }()
		assert.NoError(t, sess.SendResp(&Context{respEntry: entry}))
		sess.Close()
	})
}

func TestTCPSession_writeOutbound(t *testing.T) {
	t.Run("when session is closed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		packer := mock.NewMockPacker(ctrl)
		packer.EXPECT().Pack(gomock.Any()).AnyTimes().Return(nil, nil)

		sess := newSession(nil, &sessionOption{Packer: packer, respQueueSize: 10})
		doneLoop := make(chan struct{})
		sess.Close()
		go func() {
			sess.writeOutbound(0, 10) // should stop looping and return
			close(doneLoop)
		}()
		time.Sleep(time.Millisecond * 5)
		<-doneLoop
	})
	t.Run("when respQueue is closed", func(t *testing.T) {
		sess := newSession(nil, &sessionOption{})
		close(sess.respQueue)
		doneLoop := make(chan struct{})
		go func() {
			sess.writeOutbound(0, 10) // should stop looping and return
			close(doneLoop)
		}()
		<-doneLoop
	})
	t.Run("when response entry is nil", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		packer := mock.NewMockPacker(ctrl)
		packer.EXPECT().Pack(gomock.Any()).AnyTimes().Return(nil, nil)

		sess := newSession(nil, &sessionOption{Packer: packer, respQueueSize: 1024})
		sess.respQueue <- &Context{respEntry: nil}
		doneLoop := make(chan struct{})
		go func() {
			sess.writeOutbound(0, 10) // should stop looping and return
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

		sess := newSession(nil, &sessionOption{Packer: packer})
		done := make(chan struct{})
		go func() {
			sess.respQueue <- &Context{respEntry: entry}
			close(done)
		}()
		time.Sleep(time.Microsecond * 15)
		go sess.writeOutbound(0, 10)
		time.Sleep(time.Millisecond * 15)
		<-done
		sess.Close() // should break the write loop
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

		sess := newSession(nil, &sessionOption{Packer: packer, respQueueSize: 100})
		sess.respQueue <- &Context{respEntry: entry} // push to queue
		doneLoop := make(chan struct{})
		go func() {
			sess.writeOutbound(0, 10)
			close(doneLoop)
		}()
		time.Sleep(time.Millisecond * 5)
		sess.Close() // should break the write loop
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
		sess := newSession(p1, &sessionOption{Packer: packer})
		go func() { sess.respQueue <- &Context{respEntry: entry} }()
		go sess.writeOutbound(time.Millisecond*10, 10)
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
		sess := newSession(p1, &sessionOption{Packer: packer})
		go func() { sess.respQueue <- &Context{respEntry: entry} }()
		go sess.writeOutbound(time.Millisecond*10, 10)
		_, ok := <-sess.closed
		assert.False(t, ok)
		_ = p1.Close()
	})
	t.Run("when conn write returns fatal error", func(t *testing.T) {
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
		sess := newSession(p1, &sessionOption{Packer: packer})
		go func() { sess.respQueue <- &Context{respEntry: entry} }()
		sess.writeOutbound(0, 10) // should stop looping and return
		_, ok := <-sess.closed
		assert.False(t, ok)
	})
	t.Run("when conn write returns temporary error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		entry := &message.Entry{
			ID:   1,
			Data: []byte("test"),
		}
		packer := mock.NewMockPacker(ctrl)
		packer.EXPECT().Pack(gomock.Any()).Return([]byte("pack succeed"), nil)

		tmpErr := mock.NewMockError(ctrl)
		tmpErr.EXPECT().Timeout().Return(false)
		tmpErr.EXPECT().Temporary().Return(true)
		tmpErr.EXPECT().Error().AnyTimes().Return("some error")

		conn := mock.NewMockConn(ctrl)
		connWriteCount := 0
		conn.EXPECT().Write(gomock.Any()).Times(2).DoAndReturn(func(_ []byte) (int, error) {
			if connWriteCount == 0 {
				connWriteCount++
				return 0, tmpErr
			}
			return 0, fmt.Errorf("some fatal error")
		})

		sess := newSession(conn, &sessionOption{Packer: packer})
		go func() { sess.respQueue <- &Context{respEntry: entry} }()
		sess.writeOutbound(0, 10) // should stop looping and return
		_, ok := <-sess.closed
		assert.False(t, ok)
	})
	t.Run("when write succeed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		entry := &message.Entry{
			ID:   1,
			Data: []byte("test"),
		}
		packer := mock.NewMockPacker(ctrl)
		packer.EXPECT().Pack(gomock.Any()).Return([]byte("pack succeed"), nil)

		p1, p2 := net.Pipe()
		sess := newSession(p1, &sessionOption{Packer: packer})
		go func() {
			_ = sess.SendResp(&Context{respEntry: entry})
		}()
		done := make(chan struct{})
		go func() {
			sess.writeOutbound(0, 10)
			close(done)
		}()
		time.Sleep(time.Millisecond * 5)
		_, _ = p2.Read(make([]byte, 100))
		sess.Close()
		<-done
	})
}

func TestSession_attemptConnWrite_when_reach_last_try(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	conn := mock.NewMockConn(ctrl)
	conn.EXPECT().Write(gomock.Any()).Return(0, fmt.Errorf("some err"))

	s := newSession(conn, &sessionOption{})
	assert.Error(t, s.attemptConnWrite([]byte("whatever"), 1))
}

func TestSession_attemptConnWrite_when_err_is_not_temp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	netErr := mock.NewMockError(ctrl)
	netErr.EXPECT().Timeout().Return(false)
	netErr.EXPECT().Temporary().Return(false)

	conn := mock.NewMockConn(ctrl)
	conn.EXPECT().Write(gomock.Any()).Return(0, netErr)

	s := newSession(conn, &sessionOption{})
	assert.ErrorIs(t, s.attemptConnWrite([]byte("whatever"), 10), netErr)
}
