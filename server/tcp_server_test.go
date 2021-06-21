package server

import (
	"fmt"
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/DarthPestilane/easytcp/router"
	mock_net "github.com/DarthPestilane/easytcp/server/mock/net"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net"
	"runtime"
	"testing"
	"time"
)

func TestNewTCPServer(t *testing.T) {
	s := NewTCPServer(TCPOption{
		RWBufferSize: 0,
		ReadTimeout:  0,
		WriteTimeout: 0,
		MsgCodec:     &packet.JsonCodec{},
	})
	assert.NotNil(t, s.log)
	assert.NotNil(t, s.accepting)
	assert.Equal(t, s.msgPacker, &packet.DefaultPacker{})
	assert.Equal(t, s.msgCodec, &packet.JsonCodec{})
}

func TestTCPServer_Serve(t *testing.T) {
	goroutineNum := runtime.NumGoroutine()
	server := NewTCPServer(TCPOption{})
	go func() {
		err := server.Serve("localhost:0")
		assert.Error(t, err)
		assert.Equal(t, err, errServerStopped)
	}()
	<-server.accepting
	err := server.Stop()
	assert.NoError(t, err)
	<-time.After(time.Millisecond * 10)
	assert.Equal(t, goroutineNum, runtime.NumGoroutine()) // no goroutine leak
}

func TestTCPServer_acceptLoop(t *testing.T) {
	t.Run("when everything's fine", func(t *testing.T) {
		server := NewTCPServer(TCPOption{
			RWBufferSize: 1024,
		})
		address, err := net.ResolveTCPAddr("tcp", "localhost:0")
		assert.NoError(t, err)
		lis, err := net.ListenTCP("tcp", address)
		assert.NoError(t, err)
		server.listener = lis
		go func() {
			err := server.acceptLoop()
			assert.Error(t, err)
		}()

		<-server.accepting

		// client
		cli, err := net.Dial("tcp", lis.Addr().String())
		assert.NoError(t, err)
		assert.NoError(t, cli.Close())
		assert.NoError(t, server.Stop())
	})
	t.Run("when accept returns a non-temporary error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		server := NewTCPServer(TCPOption{})

		listen := mock_net.NewMockListener(ctrl)
		listen.EXPECT().Accept().Return(nil, fmt.Errorf("some err"))
		server.listener = listen
		go func() {
			assert.Error(t, server.acceptLoop())
		}()
		<-server.accepting
	})
	t.Run("when accept returns a temporary error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		server := NewTCPServer(TCPOption{})

		tempErr := mock_net.NewMockError(ctrl)
		tempErr.EXPECT().Error().MinTimes(1).Return("some err")
		i := 0
		tempErr.EXPECT().Temporary().MinTimes(1).DoAndReturn(func() bool {
			defer func() { i++ }()
			return i == 0 // returns true for the first time
		})

		listen := mock_net.NewMockListener(ctrl)
		listen.EXPECT().Accept().MinTimes(1).Return(nil, tempErr)
		server.listener = listen
		go func() {
			assert.Error(t, server.acceptLoop())
		}()
		<-server.accepting
		time.Sleep(time.Millisecond * 20)
	})
}

func TestTCPServer_Stop(t *testing.T) {
	server := NewTCPServer(TCPOption{})
	go func() {
		err := server.Serve("localhost:0")
		assert.Error(t, err)
		assert.Equal(t, err, errServerStopped)
	}()

	<-server.accepting

	// client
	cli, err := net.Dial("tcp", server.listener.Addr().String())
	assert.NoError(t, err)

	<-time.After(time.Millisecond * 10)

	assert.NoError(t, server.Stop()) // stop server first
	assert.NoError(t, cli.Close())
}

func TestTCPServer_handleConn(t *testing.T) {
	type TestReq struct {
		Param string
	}
	type TestResp struct {
		Success bool
	}

	// options
	codec := &packet.JsonCodec{}
	packer := &packet.DefaultPacker{}

	// server
	server := NewTCPServer(TCPOption{
		RWBufferSize: 1024,
		MsgCodec:     codec,
		MsgPacker:    packer,
	})

	// register route
	server.AddRoute(1, func(ctx *router.Context) (packet.Message, error) {
		var reqData TestReq
		assert.NoError(t, ctx.Bind(&reqData))
		assert.EqualValues(t, 1, ctx.MsgID())
		assert.Equal(t, reqData.Param, "hello test")
		return ctx.Response(2, &TestResp{Success: true})
	})
	// use middleware
	server.Use(func(next router.HandlerFunc) router.HandlerFunc {
		return func(ctx *router.Context) (packet.Message, error) {
			defer func() {
				if r := recover(); r != nil {
					assert.Fail(t, "caught panic")
				}
			}()
			return next(ctx)
		}
	})

	go func() {
		err := server.Serve("localhost:0")
		assert.Error(t, err)
		assert.Equal(t, err, errServerStopped)
	}()
	defer func() { assert.NoError(t, server.Stop()) }()

	<-server.accepting

	// client
	cli, err := net.Dial("tcp", server.listener.Addr().String())
	assert.NoError(t, err)
	defer func() { assert.NoError(t, cli.Close()) }()

	// client send msg
	reqData := &TestReq{Param: "hello test"}
	reqDataByte, err := codec.Encode(reqData)
	assert.NoError(t, err)
	msg := &packet.DefaultMsg{
		ID:   1,
		Size: uint32(len(reqDataByte)),
		Data: reqDataByte,
	}
	reqMsg, err := packer.Pack(msg)
	assert.NoError(t, err)
	_, err = cli.Write(reqMsg)
	assert.NoError(t, err)

	// client read msg
	respMsg, err := packer.Unpack(cli)
	assert.NoError(t, err)
	var respData TestResp
	assert.NoError(t, codec.Decode(respMsg.GetData(), &respData))
	assert.EqualValues(t, 2, respMsg.GetID())
	assert.True(t, respData.Success)
}
