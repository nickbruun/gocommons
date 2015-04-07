package zkutils

import (
	"errors"
)

var (
	// Node is not a match.
	ErrNodeNotMatch = errors.New("node is not a match")

	// Timeout error.
	ErrTimeout = errors.New("timeout")
)
