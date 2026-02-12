package payment

import (
	"context"

	"github.com/google/uuid"
)

// PaymentProcessor defines the interface for subscription operations
type PaymentProcessor interface {
	// CreateCustomer creates a customer in Stripe
	CreateCustomer(ctx context.Context, userId uuid.UUID, email string) (string, error)

	// SetupSubscription creates a subscription product in Stripe
	SetupSubscription(ctx context.Context, req *SetupSubscriptionReq) (*SetupSubscriptionResp, error)

	// SubscribeToProduct subscribes a customer to a product
	SubscribeToProduct(ctx context.Context, req *SubscribeRequest) (*SubscribeResponse, error)

	// GetUserSubscriptions returns all subscriptions for a customer
	GetUserSubscriptions(ctx context.Context, customerID string) (*UserSubscriptionsResponse, error)
}
