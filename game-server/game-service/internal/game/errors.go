package game

import "errors"

var (
	ErrOutOfRange = errors.New("Error when attempting to interact with door entity as it was out of range.\n")
)
