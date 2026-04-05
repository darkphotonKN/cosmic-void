package game

import "errors"

var (
	ErrOutOfRange                  = errors.New("target out of range")
	ErrEntityNotFound              = errors.New("entity not found")
	ErrComponentNotFound           = errors.New("component not found")
	ErrComponentCouldNotBeAsserted = errors.New("component assertion failed")
)
