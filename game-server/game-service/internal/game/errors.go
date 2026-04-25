package game

import "errors"

var (
	// Core game
	ErrOutOfRange                  = errors.New("target out of range")
	ErrEntityNotFound              = errors.New("entity not found")
	ErrComponentNotFound           = errors.New("component not found")
	ErrComponentCouldNotBeAsserted = errors.New("component assertion failed")

	// queue
	ErrPlayerAlreadyInQueue = errors.New("player already in queue")
)
