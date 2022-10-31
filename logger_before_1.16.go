//go:build !go1.16
// +build !go1.16

package easytcp

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

var _ Logger = &DefaultLogger{}

// Log is the instance of Logger interface.
var Log Logger = newMuteLogger()

// Logger is the generic interface for log recording.
type Logger interface {
	Errorf(format string, args ...interface{})
	Tracef(format string, args ...interface{})
}

func newLogger() *DefaultLogger {
	return &DefaultLogger{
		rawLogger: log.New(os.Stdout, "easytcp ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lmsgprefix),
	}
}

func newMuteLogger() *DefaultLogger {
	return &DefaultLogger{
		rawLogger: log.New(ioutil.Discard, "easytcp", log.LstdFlags),
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

// SetLogger sets the package logger.
func SetLogger(lg Logger) {
	Log = lg
}
