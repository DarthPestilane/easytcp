package logger

import (
	"fmt"
	"log"
	"os"
)

var Log Logger = newLogger()

type Logger interface {
	Errorf(format string, args ...interface{})
	Tracef(format string, args ...interface{})
}

type DefaultLogger struct {
	rawLogger *log.Logger
}

var _ Logger = &DefaultLogger{}

func newLogger() *DefaultLogger {
	return &DefaultLogger{
		rawLogger: log.New(os.Stdout, "easytcp ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lmsgprefix),
	}
}

func (d *DefaultLogger) Errorf(format string, args ...interface{}) {
	d.rawLogger.Printf("[ERROR] %s", fmt.Sprintf(format, args...))
}

func (d *DefaultLogger) Tracef(format string, args ...interface{}) {
	d.rawLogger.Printf("[TRACE] %s", fmt.Sprintf(format, args...))
}

type MuteLogger struct{}

var _ Logger = &MuteLogger{}

func (m *MuteLogger) Errorf(format string, args ...interface{}) {}

func (m *MuteLogger) Tracef(format string, args ...interface{}) {}
