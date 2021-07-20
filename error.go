package easytcp

import (
	"fmt"
)

// Error is a generic interface for error handling.
type Error interface {
	error
	Fatal() bool // should return true if the error is fatal, otherwise false.
}

var (
	_ error = &UnpackError{}
	_ Error = &UnpackError{}
)

// UnpackError is the error returned in packer.Unpack.
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
