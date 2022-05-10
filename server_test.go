package easytcp

import (
	"crypto/tls"
	"fmt"
	"github.com/DarthPestilane/easytcp/internal/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	s := NewServer(&ServerOption{
		ReadTimeout:   0,
		WriteTimeout:  0,
		Codec:         &JsonCodec{},
		RespQueueSize: -1,
	})
	assert.NotNil(t, s.accepting)
	assert.IsType(t, s.Packer, NewDefaultPacker())
	assert.Equal(t, s.Codec, &JsonCodec{})
	assert.Equal(t, s.respQueueSize, DefaultRespQueueSize)
	assert.NotNil(t, s.accepting)
	assert.NotNil(t, s.stopped)
}

func TestServer_Serve(t *testing.T) {
	server := NewServer(&ServerOption{})
	done := make(chan struct{})
	go func() {
		assert.ErrorIs(t, server.Serve("localhost:0"), ErrServerStopped)
		close(done)
	}()
	<-server.accepting
	time.Sleep(time.Millisecond * 5)
	err := server.Stop()
	assert.NoError(t, err)
	<-done
}

func TestServer_ServeTLS(t *testing.T) {
	server := NewServer(&ServerOption{
		SocketReadBufferSize: 123, // won't work
	})
	cert, err := tls.LoadX509KeyPair("internal/test_data/certificates/cert.pem", "internal/test_data/certificates/cert.key")
	assert.NoError(t, err)
	done := make(chan struct{})
	go func() {
		cfg := &tls.Config{
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{cert},
		}
		assert.ErrorIs(t, server.ServeTLS("localhost:0", cfg), ErrServerStopped)
		close(done)
	}()
	<-server.accepting
	time.Sleep(time.Millisecond * 5)
	err = server.Stop()
	assert.NoError(t, err)
	<-done
}

func TestServer_acceptLoop(t *testing.T) {
	t.Run("when everything's fine", func(t *testing.T) {
		server := NewServer(&ServerOption{
			SocketReadBufferSize:  1024,
			SocketWriteBufferSize: 1024,
		})
		address, err := net.ResolveTCPAddr("tcp", "localhost:0")
		assert.NoError(t, err)
		lis, err := net.ListenTCP("tcp", address)
		assert.NoError(t, err)
		server.Listener = lis
		go func() {
			err := server.acceptLoop()
			assert.Error(t, err)
		}()

		<-server.accepting

		// client
		cli, err := net.Dial("tcp", lis.Addr().String())
		assert.NoError(t, err)

		time.Sleep(time.Millisecond * 5)

		assert.NoError(t, cli.Close())
		assert.NoError(t, server.Stop())
	})
	t.Run("when accept returns a non-temporary error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		server := NewServer(&ServerOption{})

		listen := mock.NewMockListener(ctrl)
		listen.EXPECT().Accept().Return(nil, fmt.Errorf("some err"))
		server.Listener = listen
		assert.Error(t, server.acceptLoop())
	})
	t.Run("when accept returns a temporary error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		server := NewServer(&ServerOption{})

		tempErr := mock.NewMockError(ctrl)
		tempErr.EXPECT().Error().MinTimes(1).Return("some err")
		i := 0
		tempErr.EXPECT().Temporary().MinTimes(1).DoAndReturn(func() bool {
			defer func() { i++ }()
			return i == 0 // returns true for the first time
		})

		listen := mock.NewMockListener(ctrl)
		listen.EXPECT().Accept().MinTimes(1).Return(nil, tempErr)
		server.Listener = listen
		assert.Error(t, server.acceptLoop())
	})
	t.Run("when server is stopped", func(t *testing.T) {
		server := NewServer(&ServerOption{
			SocketReadBufferSize:  1024,
			SocketWriteBufferSize: 1024,
		})
		address, err := net.ResolveTCPAddr("tcp", "localhost:0")
		assert.NoError(t, err)
		lis, err := net.ListenTCP("tcp", address)
		assert.NoError(t, err)
		server.Listener = lis
		assert.NoError(t, server.Stop())
		assert.ErrorIs(t, server.acceptLoop(), ErrServerStopped)
	})
}

func TestServer_Stop(t *testing.T) {
	server := NewServer(&ServerOption{})
	go func() {
		err := server.Serve("localhost:0")
		assert.Error(t, err)
		assert.Equal(t, err, ErrServerStopped)
	}()

	<-server.accepting

	// client
	cli, err := net.Dial("tcp", server.Listener.Addr().String())
	assert.NoError(t, err)

	time.Sleep(time.Millisecond * 5)

	assert.NoError(t, server.Stop()) // stop server first
	assert.NoError(t, cli.Close())
}

func TestServer_handleConn(t *testing.T) {
	type TestReq struct {
		Param string
	}
	type TestResp struct {
		Success bool
	}

	// options
	codec := &JsonCodec{}
	packer := NewDefaultPacker()

	// server
	server := NewServer(&ServerOption{
		SocketReadBufferSize:  1,
		SocketWriteBufferSize: 1,
		SocketSendDelay:       true,
		Codec:                 codec,
		Packer:                packer,
		RespQueueSize:         -1,
		AsyncRouter:           true,
	})

	// hooks
	server.OnSessionCreate = func(sess Session) {
		fmt.Printf("session created | id: %s\n", sess.ID())
	}
	server.OnSessionClose = func(sess Session) {
		fmt.Printf("session closed | id: %s\n", sess.ID())
	}

	// register route
	server.AddRoute(1, func(ctx Context) {
		var reqData TestReq
		assert.NoError(t, ctx.Bind(&reqData))
		assert.EqualValues(t, 1, ctx.Request().ID())
		assert.Equal(t, reqData.Param, "hello test")
		ctx.MustSetResponse(2, &TestResp{Success: true})
	})
	// use middleware
	server.Use(func(next HandlerFunc) HandlerFunc {
		return func(ctx Context) {
			defer func() {
				if r := recover(); r != nil {
					assert.Fail(t, "caught panic")
				}
			}()
			next(ctx)
		}
	})

	go func() {
		err := server.Serve("localhost:0")
		assert.Error(t, err)
		assert.Equal(t, err, ErrServerStopped)
	}()
	defer func() { assert.NoError(t, server.Stop()) }()

	<-server.accepting

	// client
	cli, err := net.Dial("tcp", server.Listener.Addr().String())
	assert.NoError(t, err)
	defer func() { assert.NoError(t, cli.Close()) }()

	// client send msg
	reqData := &TestReq{Param: "hello test"}
	reqDataByte, err := codec.Encode(reqData)
	assert.NoError(t, err)
	reqMsg, err := packer.Pack(NewMessage(1, reqDataByte))
	assert.NoError(t, err)
	_, err = cli.Write(reqMsg)
	assert.NoError(t, err)

	// client read msg
	respMsg, err := packer.Unpack(cli)
	assert.NoError(t, err)
	var respData TestResp
	assert.NoError(t, codec.Decode(respMsg.Data(), &respData))
	assert.EqualValues(t, 2, respMsg.ID())
	assert.True(t, respData.Success)
}

func TestServer_NotFoundHandler(t *testing.T) {
	// server
	server := NewServer(&ServerOption{
		Packer: NewDefaultPacker(),
	})
	server.NotFoundHandler(func(ctx Context) {
		ctx.SetResponseMessage(NewMessage(101, []byte("handler not found")))
	})
	go func() {
		err := server.Serve(":0")
		assert.Equal(t, err, ErrServerStopped)
	}()

	<-server.accepting

	// client
	cli, err := net.Dial("tcp", server.Listener.Addr().String())
	assert.NoError(t, err)
	defer func() { assert.NoError(t, cli.Close()) }()

	// send msg
	reqBytes, err := server.Packer.Pack(NewMessage(1, []byte("test")))
	assert.NoError(t, err)
	_, err = cli.Write(reqBytes)
	assert.NoError(t, err)

	// read msg
	reqMsg, err := server.Packer.Unpack(cli)
	assert.NoError(t, err)
	assert.EqualValues(t, reqMsg.ID(), 101)
	assert.Equal(t, reqMsg.Data(), []byte("handler not found"))
}
