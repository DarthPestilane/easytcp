package easytcp

import (
	"github.com/DarthPestilane/easytcp/logger"
	"github.com/DarthPestilane/easytcp/server"
	"github.com/sirupsen/logrus"
)

func SetLogger(log *logrus.Logger) {
	logger.Default = log
}

func NewTcpServer(opt server.TCPOption) *server.TCPServer {
	return server.NewTCPServer(opt)
}

func NewUdpServer(opt server.UDPOption) *server.UDPServer {
	return server.NewUDPServer(opt)
}
