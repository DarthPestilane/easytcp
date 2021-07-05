package server

import (
	"fmt"
	"net"
)

//go:generate mockgen -destination mock/net/net.go -package mock_net net Listener,Error

// ErrServerStopped is used when server stopped.
var ErrServerStopped = fmt.Errorf("server stopped")

func isStopped(stopChan <-chan struct{}) bool {
	select {
	case <-stopChan:
		return true
	default:
		return false
	}
}

func isTempErr(err error) bool {
	ne, ok := err.(net.Error)
	return ok && ne.Temporary()
}
