package notary

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNotFoundError(t *testing.T) {
	assert := assert.New(t)

	err := ErrItemNotFound("1234")

	assert.True(errors.Is(err, ErrNotFound), "expected a ErrNotFound")
}
