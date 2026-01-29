package payment

import (
	"time"

	"github.com/google/uuid"
)

// Subscription Entity
type Subscription struct {
	ID                   uuid.UUID  `db:"id" json:"id"`
	UserID               uuid.UUID  `db:"user_id" json:"userId"`
	StripeCustomerID     string     `db:"stripe_customer_id" json:"stripeCustomerId"`
	StripeSubscriptionID string     `db:"stripe_subscription_id" json:"stripeSubscriptionId"`
	StripePriceID        string     `db:"stripe_price_id" json:"stripePriceId"`
	Status               string     `db:"status" json:"status"`
	CurrentPeriodStart   time.Time  `db:"current_period_start" json:"currentPeriodStart"`
	CurrentPeriodEnd     time.Time  `db:"current_period_end" json:"currentPeriodEnd"`
	CancelAtPeriodEnd    bool       `db:"cancel_at_period_end" json:"cancelAtPeriodEnd"`
	CanceledAt           *time.Time `db:"canceled_at" json:"canceledAt"`
	CreatedAt            time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt            time.Time  `db:"updated_at" json:"updatedAt"`
}

// --- Request / Response ---

// Create Customer
type CreateCustomerReq struct {
	Email string `json:"email"`
}

type CreateCustomerResponse struct {
	CustomerID string `json:"customerId"`
}

// Setup Subscription Product
type SetupSubscriptionReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int64  `json:"price"` // in cents
}

type SetupSubscriptionResp struct {
	ProductID string `json:"productId"`
	PriceID   string `json:"priceId"`
}

// Subscribe to Product
type SubscribeRequest struct {
	ProductID  string `json:"productId"`
	CustomerID string `json:"customerId"`
}

type SubscribeResponse struct {
	SubscriptionID string `json:"subscriptionId"`
	ClientSecret   string `json:"clientSecret"`
	Status         string `json:"status"`
}

// Get User Subscriptions
type UserSubscriptionInfo struct {
	SubscriptionID string `json:"subscriptionId"`
	ProductID      string `json:"productId"`
	PriceID        string `json:"priceId"`
	Status         string `json:"status"`
	CurrentPeriodEnd int64 `json:"currentPeriodEnd"`
}

type UserSubscriptionsResponse struct {
	Subscriptions []UserSubscriptionInfo `json:"subscriptions"`
}
