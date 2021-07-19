package easytcp

import (
	"fmt"
)

type Error interface {
	error
	Fatal() bool
}

type UnpackError struct {
	Err error
}

func (pe *UnpackError) Error() string {
	return pe.Err.Error()
}

func (pe *UnpackError) Fatal() bool {
	return true
}

// ErrServerStopped is used when server stopped.
var ErrServerStopped = fmt.Errorf("server stopped")
