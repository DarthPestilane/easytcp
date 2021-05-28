package server

import (
	"github.com/stretchr/testify/assert"
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
