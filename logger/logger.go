package logger

import (
	"fmt"
	"log"
	"os"
)

// Log is the instance of Logger interface.
var Log Logger = newLogger()

// Logger is the generic interface for log recording.
type Logger interface {
	Errorf(format string, args ...interface{})
	Tracef(format string, args ...interface{})
}

// DefaultLogger is the default logger instance for this package.
// DefaultLogger uses the built-in log.Logger.
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

// MuteLogger is the empty logger instance.
type MuteLogger struct{}

var _ Logger = &MuteLogger{}

func (m *MuteLogger) Errorf(format string, args ...interface{}) {}

func (m *MuteLogger) Tracef(format string, args ...interface{}) {}
