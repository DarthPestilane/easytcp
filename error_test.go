package easytcp

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUnpackError(t *testing.T) {
	err := fmt.Errorf("something fatal")
	ue := &UnpackError{Err: err}
	assert.ErrorIs(t, err, ue.Err)
	assert.Equal(t, err.Error(), ue.Error())
	assert.True(t, ue.Fatal())
}
