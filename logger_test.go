package easytcp

import (
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

func newLogger() *DefaultLogger {
	return &DefaultLogger{
		rawLogger: log.New(os.Stdout, "easytcp ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lmsgprefix),
	}
}

func TestDefaultLogger_Errorf(t *testing.T) {
	lg := newLogger()
	lg.Errorf("err: %s", "some error")
}

func TestDefaultLogger_Tracef(t *testing.T) {
	lg := newLogger()
	lg.Tracef("some trace info: %s", "here")
}

func TestSetLogger(t *testing.T) {
	lg := &mutedLogger{}
	SetLogger(lg)
	assert.Equal(t, Log(), lg)
}
