package notary

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrItemNotFound = func(item string) error { return fmt.Errorf("%q: %w", item, ErrNotFound) }
	ErrInvalidID    = errors.New("invalid id")
)
