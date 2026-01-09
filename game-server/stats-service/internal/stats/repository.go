package stats

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Repository interface {
	GetPlayerStats(ctx context.Context, playerID uuid.UUID) (*PlayerStats, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) GetPlayerStats(ctx context.Context, playerID uuid.UUID) (*PlayerStats, error) {
	return nil, fmt.Errorf("not implemented")
}