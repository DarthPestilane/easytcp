package server

import (
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
	"time"
)

func TestNewUdp(t *testing.T) {
	u := NewUdp(UdpOption{})
	assert.NotNil(t, u.log)
	assert.NotNil(t, u.accepting)
	assert.Equal(t, u.msgPacker, &packet.DefaultPacker{})
	assert.Equal(t, u.msgCodec, &packet.StringCodec{})
	assert.Equal(t, u.maxBufferSize, 1024)
}

func TestUdpServer_Serve(t *testing.T) {
	t.Run("when addr is invalid", func(t *testing.T) {
		server := NewUdp(UdpOption{RWBufferSize: 1024})
		assert.Error(t, server.Serve("invalid"))

		// when address is in use
		go func() {
			_ = server.Serve("localhost:0")
		}()
		<-server.accepting
		server2 := NewUdp(UdpOption{RWBufferSize: 1024})
		assert.Error(t, server2.Serve(server.conn.LocalAddr().String()))

	})
	t.Run("when ReadFromUDP failed", func(t *testing.T) {
		server := NewUdp(UdpOption{})
		go func() {
			assert.Error(t, server.Serve("localhost:0"))
		}()
		<-server.accepting
		_ = server.conn.Close()
	})
	t.Run("when ReadFromUDP succeed", func(t *testing.T) {
		server := NewUdp(UdpOption{})
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

func TestUdpServer_Stop(t *testing.T) {
	server := NewUdp(UdpOption{})
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
