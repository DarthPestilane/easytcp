package server

import (
	"fmt"
	"github.com/DarthPestilane/easytcp/router"
	"net"
)

//go:generate mockgen -destination mock/net/net.go -package mock_net net Listener,Error

// Server is a generic network server.
type Server interface {
	// Serve starts to serving at the addr.
	// Returns error when error occurred.
	Serve(addr string) error

	// Stop stops the server from serving.
	// All the goroutines created from Server should be exit.
	// Returns error when error occurred.
	Stop() error

	// Use registers a list of global middlewares.
	Use(middlewares ...router.MiddlewareFunc)

	// AddRoute registers a handler and middlewares for the message matches msgID.
	AddRoute(msgID uint, handler router.HandlerFunc, middlewares ...router.MiddlewareFunc)
}

var errServerStopped = fmt.Errorf("server stopped")

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
