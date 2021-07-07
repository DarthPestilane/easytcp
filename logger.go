package easytcp

import (
	"fmt"
	"log"
	"os"
)

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

// Log is the instance of Logger interface.
var Log Logger = newLogger()

func newLogger() *DefaultLogger {
	return &DefaultLogger{
		rawLogger: log.New(os.Stdout, "easytcp ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lmsgprefix),
	}
}

// Errorf implements Logger Errorf method.
func (d *DefaultLogger) Errorf(format string, args ...interface{}) {
	d.rawLogger.Printf("[ERROR] %s", fmt.Sprintf(format, args...))
}

// Tracef implements Logger Tracef method.
func (d *DefaultLogger) Tracef(format string, args ...interface{}) {
	d.rawLogger.Printf("[TRACE] %s", fmt.Sprintf(format, args...))
}

// MuteLogger is the empty logger instance.
type MuteLogger struct{}

var _ Logger = &MuteLogger{}

// Errorf is an empty implementation to Logger Errorf method.
func (m *MuteLogger) Errorf(format string, args ...interface{}) {}

// Tracef is an empty implementation to Logger Tracef method.
func (m *MuteLogger) Tracef(format string, args ...interface{}) {}

// SetLogger sets the package logger.
func SetLogger(lg Logger) {
	Log = lg
}
