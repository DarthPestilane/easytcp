package server

import (
	"github.com/DarthPestilane/easytcp/router"
)

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

	// AddRoute registers a handler for the message matches msgId.
	// Also registers middlewares.
	AddRoute(msgId uint, handler router.HandlerFunc, middlewares ...router.MiddlewareFunc)
}
