package easytcp

import (
	"github.com/DarthPestilane/easytcp/logger"
	"github.com/DarthPestilane/easytcp/server"
	"github.com/sirupsen/logrus"
)

// SetLogger sets the logger for package.
func SetLogger(log *logrus.Logger) {
	logger.Default = log
}

// NewTCPServer creates a new server.TCPServer according to opt.
func NewTCPServer(opt server.TCPOption) *server.TCPServer {
	return server.NewTCPServer(opt)
}

// NewUDPServer creates a new server.UDPServer according to opt.
func NewUDPServer(opt server.UDPOption) *server.UDPServer {
	return server.NewUDPServer(opt)
}
