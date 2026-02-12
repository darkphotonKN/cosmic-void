package payment

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Repository interface {
	UpsertSubscription(ctx context.Context, sub *Subscription) error
	GetSubscriptionsByUserID(ctx context.Context, userID uuid.UUID) ([]Subscription, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) UpsertSubscription(ctx context.Context, sub *Subscription) error {
	query := `
		INSERT INTO subscriptions (
			id, user_id, stripe_customer_id, stripe_subscription_id, stripe_price_id,
			status, current_period_start, current_period_end, cancel_at_period_end, canceled_at
		) VALUES (
			:id, :user_id, :stripe_customer_id, :stripe_subscription_id, :stripe_price_id,
			:status, :current_period_start, :current_period_end, :cancel_at_period_end, :canceled_at
		)
		ON CONFLICT (stripe_subscription_id) DO UPDATE SET
			status = EXCLUDED.status,
			stripe_price_id = EXCLUDED.stripe_price_id,
			current_period_start = EXCLUDED.current_period_start,
			current_period_end = EXCLUDED.current_period_end,
			cancel_at_period_end = EXCLUDED.cancel_at_period_end,
			canceled_at = EXCLUDED.canceled_at,
			updated_at = NOW()
	`

	_, err := r.db.NamedExecContext(ctx, query, sub)
	if err != nil {
		return fmt.Errorf("failed to upsert subscription: %w", err)
	}
	return nil
}

func (r *repository) GetSubscriptionsByUserID(ctx context.Context, userID uuid.UUID) ([]Subscription, error) {
	var subs []Subscription
	query := `SELECT * FROM subscriptions WHERE user_id = $1 ORDER BY created_at DESC`

	err := r.db.SelectContext(ctx, &subs, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions: %w", err)
	}
	return subs, nil
}
