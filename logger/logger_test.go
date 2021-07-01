package logger

import (
	"testing"
)

func TestDefaultLogger_Errorf(t *testing.T) {
	lg := newLogger()
	lg.Errorf("err: %s", "some error")
}

func TestDefaultLogger_Tracef(t *testing.T) {
	lg := newLogger()
	lg.Tracef("some trace info: %s", "here")
}
