package auth

import (
	"time"
)

// Auth represents a basic auth entity
type Auth struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// AuthCreate represents the data needed to create a new auth
type AuthCreate struct {
	Name string `json:"name" validate:"required"`
}

// AuthUpdate represents the data needed to update an existing auth
type AuthUpdate struct {
	Name string `json:"name" validate:"required"`
}

