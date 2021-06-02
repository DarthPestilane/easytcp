package server

import (
	"github.com/stretchr/testify/assert"
	"net"
	"runtime"
	"testing"
	"time"
)

func TestTcpServer_Serve(t *testing.T) {
	goroutineNum := runtime.NumGoroutine()
	server := NewTcp(TcpOption{})
	go func() {
		err := server.Serve("localhost:0")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "accept err")
	}()
	<-time.After(time.Millisecond * 10)
	err := server.Stop()
	assert.NoError(t, err)
	<-time.After(time.Millisecond * 10)
	assert.Equal(t, goroutineNum, runtime.NumGoroutine()) // no goroutine leak
}

func TestTcpServer_acceptLoop(t *testing.T) {
	server := NewTcp(TcpOption{
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

	<-time.After(time.Millisecond * 10)

	// client
	cli, err := net.Dial("tcp", lis.Addr().String())
	assert.NoError(t, err)
	assert.NoError(t, cli.Close())
	assert.NoError(t, server.Stop())
}

func TestTcpServer_Stop(t *testing.T) {
	server := NewTcp(TcpOption{})
	go func() {
		err := server.Serve("localhost:0")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "accept err")
	}()

	<-time.After(time.Millisecond * 10)

	// client
	cli, err := net.Dial("tcp", server.listener.Addr().String())
	assert.NoError(t, err)

	<-time.After(time.Millisecond * 10)

	// close server and client
	assert.NoError(t, server.Stop())
	assert.NoError(t, cli.Close())
}

// func TestTcpServer_handleConn(t *testing.T) {
// 	// options
// 	codec := &packet.JsonCodec{}
// 	packer := &packet.DefaultPacker{}
//
// 	type TestReq struct {
// 		Param string
// 	}
// 	type TestResp struct {
// 		Success bool
// 	}
//
// 	// server
// 	server := NewTcp(TcpOption{
// 		RWBufferSize: 1024,
// 		MsgCodec:     codec,
// 		MsgPacker:    packer,
// 	})
//
// 	// register route
// 	router.Instance().Register(1, func(s session.Session, req *packet.Request) (*packet.Response, error) {
// 		var reqData TestReq
// 		assert.NoError(t, s.MsgCodec().Decode(req.RawData, &reqData))
// 		assert.Equal(t, reqData.Param, "hello test")
// 		resp := &packet.Response{
// 			Id:   2,
// 			Data: &TestResp{Success: true},
// 		}
// 		return resp, nil
// 	})
//
// 	go func() {
// 		err := server.Serve("localhost:0")
// 		assert.Error(t, err)
// 		assert.Contains(t, err.Error(), "accept err")
// 	}()
//
// 	<-time.After(time.Millisecond * 10)
//
// 	// client
// 	cli, err := net.Dial("tcp", server.listener.Addr().String())
// 	assert.NoError(t, err)
// 	// encode msg
// 	reqData := &TestReq{Param: "hello test"}
// 	reqDataByte, err := codec.Encode(reqData)
// 	assert.NoError(t, err)
// 	reqMsg, err := packer.Pack(1, reqDataByte)
// 	assert.NoError(t, err)
// 	_, err = cli.Write(reqMsg)
// 	assert.NoError(t, err)
// }
