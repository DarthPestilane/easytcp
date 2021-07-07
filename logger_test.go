package easytcp

import (
	"github.com/stretchr/testify/assert"
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

func TestSetLogger(t *testing.T) {
	lg := &MuteLogger{}
	SetLogger(lg)
	assert.Equal(t, Log, lg)
}
