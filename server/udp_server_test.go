package server

import (
	"bytes"
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/DarthPestilane/easytcp/router"
	"github.com/DarthPestilane/easytcp/session"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
	"time"
)

func TestNewUDPServer(t *testing.T) {
	u := NewUDPServer(UDPOption{})
	assert.NotNil(t, u.log)
	assert.NotNil(t, u.accepting)
	assert.Equal(t, u.msgPacker, &packet.DefaultPacker{})
	assert.Equal(t, u.msgCodec, &packet.StringCodec{})
	assert.Equal(t, u.maxBufferSize, 1024)
}

func TestUDPServer_Serve(t *testing.T) {
	t.Run("when addr is invalid", func(t *testing.T) {
		server := NewUDPServer(UDPOption{RWBufferSize: 1024})
		assert.Error(t, server.Serve("invalid"))

		// when address is in use
		go func() {
			_ = server.Serve("localhost:0")
		}()
		<-server.accepting
		server2 := NewUDPServer(UDPOption{RWBufferSize: 1024})
		assert.Error(t, server2.Serve(server.conn.LocalAddr().String()))

	})
	t.Run("when ReadFromUDP failed", func(t *testing.T) {
		server := NewUDPServer(UDPOption{})
		go func() {
			assert.Error(t, server.Serve("localhost:0"))
		}()
		<-server.accepting
		_ = server.conn.Close()
	})
	t.Run("when ReadFromUDP succeed", func(t *testing.T) {
		server := NewUDPServer(UDPOption{})
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
	server := NewUDPServer(UDPOption{})
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
	codec := &packet.StringCodec{}
	packer := &packet.DefaultPacker{}

	server := NewUDPServer(UDPOption{
		MsgCodec:  codec,
		MsgPacker: packer,
	})
	// use middleware
	server.Use(func(next router.HandlerFunc) router.HandlerFunc {
		return func(s session.Session, req *packet.Request) (*packet.Response, error) {
			defer func() {
				if r := recover(); r != nil {
					assert.Fail(t, "caught panic")
				}
			}()
			return next(s, req)
		}
	})
	// register route
	server.AddRoute(1, func(s session.Session, req *packet.Request) (*packet.Response, error) {
		return &packet.Response{
			ID:   2,
			Data: "test-resp",
		}, nil
	})
	go func() {
		assert.Error(t, server.Serve("localhost:0"))
	}()

	<-server.accepting

	// client
	cli, err := net.Dial("udp", server.conn.LocalAddr().String())
	assert.NoError(t, err)

	// send req
	b, err := codec.Encode("test-req")
	assert.NoError(t, err)
	msg, err := packer.Pack(1, b)
	assert.NoError(t, err)
	_, err = cli.Write(msg)
	assert.NoError(t, err)

	// receive resp
	buff := make([]byte, 128)
	n, err := cli.Read(buff)
	assert.NoError(t, err)
	respMsg, err := packer.Unpack(bytes.NewReader(buff[:n]))
	assert.NoError(t, err)
	assert.EqualValues(t, 2, respMsg.GetID())
	var resp string
	assert.NoError(t, codec.Decode(respMsg.GetData(), &resp))
	assert.Equal(t, resp, "test-resp")

	assert.NoError(t, server.Stop())
	assert.NoError(t, cli.Close())
}
