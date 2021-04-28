package logger

import (
	"github.com/sirupsen/logrus"
)

var Default *logrus.Logger

func init() {
	rawLog := *logrus.StandardLogger()
	rawLog.SetLevel(logrus.TraceLevel)
	Default = &rawLog
}
