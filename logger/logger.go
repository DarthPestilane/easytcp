package logger

import (
	"github.com/sirupsen/logrus"
)

var Default *logrus.Logger

func init() {
	Default = logrus.New()
	Default.SetLevel(logrus.TraceLevel)
	Default.SetFormatter(NewTextFormatter())
}
