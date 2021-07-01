package server

import (
	"bytes"
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/DarthPestilane/easytcp/router"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
	"time"
)

func TestNewUDPServer(t *testing.T) {
	u := NewUDPServer(&UDPOption{
		MsgCodec: &packet.JsonCodec{},
	})
	assert.NotNil(t, u.accepting)
	assert.NotNil(t, u.stopped)
	assert.Equal(t, u.msgPacker, &packet.DefaultPacker{})
	assert.Equal(t, u.msgCodec, &packet.JsonCodec{})
	assert.Equal(t, u.maxBufferSize, 1024)
}

func TestUDPServer_Serve(t *testing.T) {
	t.Run("when addr is invalid", func(t *testing.T) {
		server := NewUDPServer(&UDPOption{SocketRWBufferSize: 1024})
		assert.Error(t, server.Serve("invalid"))

		// when address is in use
		go func() {
			_ = server.Serve("localhost:0")
		}()
		<-server.accepting
		server2 := NewUDPServer(&UDPOption{SocketRWBufferSize: 1024})
		assert.Error(t, server2.Serve(server.conn.LocalAddr().String()))

	})
	t.Run("when ReadFromUDP failed", func(t *testing.T) {
		server := NewUDPServer(&UDPOption{})
		go func() {
			assert.Error(t, server.Serve("localhost:0"))
		}()
		<-server.accepting
		_ = server.conn.Close()
	})
	t.Run("when ReadFromUDP succeed", func(t *testing.T) {
		server := NewUDPServer(&UDPOption{})
		go func() {
			assert.Error(t, server.Serve("localhost:0"))
		}()
		<-server.accepting

		// client
		client, err := net.Dial("udp", server.conn.LocalAddr().String())
		assert.NoError(t, err)

		_, err = client.Write([]byte("test"))
		assert.NoError(t, err)

		assert.NoError(t, server.Stop())
	})
}

func TestUDPServer_Stop(t *testing.T) {
	server := NewUDPServer(&UDPOption{})
	go func() {
		assert.Error(t, server.Serve("localhost:0"))
	}()

	<-server.accepting

	// client
	cli, err := net.Dial("udp", server.conn.LocalAddr().String())
	assert.NoError(t, err)

	<-time.After(time.Millisecond * 10)

	assert.NoError(t, server.Stop()) // stop server first
	assert.NoError(t, cli.Close())
}

func TestUDPServer_handleIncomingMsg(t *testing.T) {
	packer := &packet.DefaultPacker{}

	server := NewUDPServer(&UDPOption{
		MsgPacker: packer,
	})
	// use middleware
	server.Use(func(next router.HandlerFunc) router.HandlerFunc {
		return func(ctx *router.Context) (*packet.MessageEntry, error) {
			defer func() {
				if r := recover(); r != nil {
					assert.Fail(t, "caught panic")
				}
			}()
			return next(ctx)
		}
	})
	// register route
	server.AddRoute(1, func(ctx *router.Context) (*packet.MessageEntry, error) {
		return ctx.Response(2, "test-resp")
	})
	go func() {
		assert.Error(t, server.Serve("localhost:0"))
	}()

	<-server.accepting

	// client
	cli, err := net.Dial("udp", server.conn.LocalAddr().String())
	assert.NoError(t, err)

	// send req
	b := []byte("test-req")
	assert.NoError(t, err)

	msg, err := packer.Pack(&packet.MessageEntry{
		ID:   1,
		Data: b,
	})
	assert.NoError(t, err)
	_, err = cli.Write(msg)
	assert.NoError(t, err)

	// receive resp
	buff := make([]byte, 128)
	n, err := cli.Read(buff)
	assert.NoError(t, err)
	respMsg, err := packer.Unpack(bytes.NewReader(buff[:n]))
	assert.NoError(t, err)
	assert.EqualValues(t, 2, respMsg.ID)
	assert.Equal(t, respMsg.Data, []byte("test-resp"))

	assert.NoError(t, server.Stop())
	assert.NoError(t, cli.Close())
}
