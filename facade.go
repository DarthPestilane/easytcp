package easytcp

import (
	"github.com/DarthPestilane/easytcp/logger"
	"github.com/DarthPestilane/easytcp/server"
	"github.com/sirupsen/logrus"
)

func SetLogger(log *logrus.Logger) {
	logger.Default = log
}

func NewTcp(opt server.TcpOption) *server.TcpServer {
	return server.NewTcp(opt)
}

func NewUdp(opt server.UdpOption) *server.UdpServer {
	return server.NewUdp(opt)
}
