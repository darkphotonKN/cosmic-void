package types

import "github.com/google/uuid"

type Player struct {
	ID       uuid.UUID
	Username string
}
