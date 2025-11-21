package game

import "github.com/google/uuid"

type Player struct {
	id       uuid.UUID
	username string
}
