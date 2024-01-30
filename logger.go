package easytcp

import (
	"fmt"
	"io"
	"log"
)

var _ Logger = &DefaultLogger{}

// _log is the instance of Logger interface.
var _log Logger = newDiscardLogger()

// Logger is the generic interface for log recording.
type Logger interface {
	Errorf(format string, args ...interface{})
	Tracef(format string, args ...interface{})
}

func newDiscardLogger() *DefaultLogger {
	return &DefaultLogger{
		rawLogger: log.New(io.Discard, "easytcp", log.LstdFlags),
	}
}

// DefaultLogger is the default logger instance for this package.
// DefaultLogger uses the built-in log.Logger.
type DefaultLogger struct {
	rawLogger *log.Logger
}

// Errorf implements Logger Errorf method.
func (d *DefaultLogger) Errorf(format string, args ...interface{}) {
	d.rawLogger.Printf("[ERROR] %s", fmt.Sprintf(format, args...))
}

// Tracef implements Logger Tracef method.
func (d *DefaultLogger) Tracef(format string, args ...interface{}) {
	d.rawLogger.Printf("[TRACE] %s", fmt.Sprintf(format, args...))
}

// Log returns the package logger.
func Log() Logger {
	return _log
}

// SetLogger sets the package logger.
func SetLogger(lg Logger) {
	_log = lg
}
