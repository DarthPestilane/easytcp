package util

import (
	"io"
	"strings"
)

func IsEOF(err error) bool {
	return err == io.EOF || strings.Contains(strings.ToLower(err.Error()), "connection reset by peer")
}
