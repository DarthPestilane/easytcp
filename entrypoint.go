package easytcp

import (
	"github.com/DarthPestilane/easytcp/logger"
	"github.com/sirupsen/logrus"
)

func SetLogger(l *logrus.Logger) {
	logger.Default = l
}
