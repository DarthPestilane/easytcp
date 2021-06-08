package server

import (
	"github.com/DarthPestilane/easytcp/router"
)

type Server interface {
	Serve(addr string) error
	Stop() error
	Use(middlewares ...router.MiddlewareFunc)
	AddRoute(msgId uint, handler router.HandlerFunc, middlewares ...router.MiddlewareFunc)
}
