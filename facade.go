package easytcp

import (
	"github.com/DarthPestilane/easytcp/logger"
	"github.com/DarthPestilane/easytcp/server"
	"github.com/sirupsen/logrus"
)

func SetLogger(log *logrus.Logger) {
	logger.Default = log
}

func NewTcpServer(opt server.TcpOption) *server.TcpServer {
	return server.NewTcpServer(opt)
}

func NewUdpServer(opt server.UdpOption) *server.UdpServer {
	return server.NewUdpServer(opt)
}
