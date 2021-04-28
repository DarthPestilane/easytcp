package easytcp

import (
	"github.com/DarthPestilane/easytcp/core"
	"github.com/DarthPestilane/easytcp/logger"
	"github.com/sirupsen/logrus"
)

func SetLogger(l *logrus.Logger) {
	logger.Default = l
}

func NewServer(addr string, port int) *core.Server {
	return core.NewServer(addr, port)
}
